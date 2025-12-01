import Foundation

// MARK: - Config
struct Config {
    let baseURL: URL
    let intervalSeconds: Int
    let deviceID: String

    static func load() throws -> Config {
        let fm = FileManager.default
        let home = fm.homeDirectoryForCurrentUser
        let stateDir = home.appendingPathComponent(".raqeem", isDirectory: true)
        let idFile = stateDir.appendingPathComponent("device_id.json")

        let env = ProcessInfo.processInfo.environment
        // Backend endpoint as plain text "ip:port". Resolution order:
        // 1) DEVICES_BACKEND_ADDR env (e.g., 127.0.0.1:30080)
        // 2) ~/.raqeem/backend_addr.txt (first line like 10.0.0.5:30080)
        // 3) default 127.0.0.1:30080
        var addr = env["DEVICES_BACKEND_ADDR"]
        if addr == nil {
            let addrFile = stateDir.appendingPathComponent("backend_addr.txt")
            if fm.fileExists(atPath: addrFile.path),
               let s = try? String(contentsOf: addrFile).trimmingCharacters(in: .whitespacesAndNewlines),
               !s.isEmpty {
                addr = s
            }
        }
        let effectiveAddr = (addr?.isEmpty == false) ? addr! : "127.0.0.1:30080"
        guard let baseURL = URL(string: "http://\(effectiveAddr)/api/v1") else { throw NSError(domain: "cfg", code: 1) }
        let intervalSeconds = Int(env["INTERVAL_SECONDS"] ?? "30") ?? 30

        var deviceID = env["DEVICE_ID"]
        if deviceID == nil {
            // Read or generate persisted device ID
            if fm.fileExists(atPath: idFile.path) {
                if let data = try? Data(contentsOf: idFile),
                   let obj = try? JSONSerialization.jsonObject(with: data) as? [String: Any],
                   let id = obj["deviceid"] as? String {
                    deviceID = id
                }
            }
            if deviceID == nil {
                let newID = UUID().uuidString
                try? fm.createDirectory(at: stateDir, withIntermediateDirectories: true)
                let data = try JSONSerialization.data(withJSONObject: ["deviceid": newID], options: [.prettyPrinted])
                try? data.write(to: idFile)
                deviceID = newID
            }
        }
        return Config(baseURL: baseURL, intervalSeconds: intervalSeconds, deviceID: deviceID!)
    }
}

// MARK: - HTTP helpers
struct HTTP {
    static func postJSON(url: URL, body: Any) async throws -> (Int, Data) {
        var req = URLRequest(url: url)
        req.httpMethod = "POST"
        req.setValue("application/json", forHTTPHeaderField: "Content-Type")
        req.httpBody = try JSONSerialization.data(withJSONObject: body, options: [])

        let (data, resp) = try await URLSession.shared.data(for: req)
        let status = (resp as? HTTPURLResponse)?.statusCode ?? -1
        return (status, data)
    }

    static func getJSON(url: URL) async throws -> (Int, Data) {
        var req = URLRequest(url: url)
        req.httpMethod = "GET"
        let (data, resp) = try await URLSession.shared.data(for: req)
        let status = (resp as? HTTPURLResponse)?.statusCode ?? -1
        return (status, data)
    }
}

// MARK: - System info & metrics
func currentUser() -> String {
    return NSUserName()
}

func hostName() -> String {
    Host.current().localizedName ?? ProcessInfo.processInfo.hostName
}

func osName() -> String { "macOS" }

func runShell(_ cmd: String, _ args: [String]) -> String {
    let p = Process()
    p.executableURL = URL(fileURLWithPath: "/usr/bin/env")
    p.arguments = [cmd] + args
    let out = Pipe()
    p.standardOutput = out
    p.standardError = Pipe()
    do { try p.run() } catch { return "" }
    p.waitUntilExit()
    let data = out.fileHandleForReading.readDataToEndOfFile()
    return String(data: data, encoding: .utf8) ?? ""
}

struct Metrics {
    var cpuUsage: Double?
    var memTotalMB: Int?
    var memUsedMB: Int?
    var diskTotalMB: Int?
    var diskUsedMB: Int?
    var netIn: Int?
    var netOut: Int?
}

func collectMemory() -> (totalMB: Int?, usedMB: Int?) {
    // Total via ProcessInfo; Used via vm_stat parsing
    let totalBytes = ProcessInfo.processInfo.physicalMemory
    let totalMB = Int(totalBytes / 1_048_576)
    let vm = runShell("vm_stat", [])
    // Parse pages statistics
    var pageSize = 4096
    if let m = vm.firstMatch(of: /page size of (\d+) bytes/), let size = Int(m.1) { pageSize = size }
    func pages(_ key: String) -> Int64 {
        if let m = vm.firstMatch(of: Regex("^\\s*\\Q\(pageSize\)\\E?")) { _ = m }
        let lines = vm.split(separator: "\n").map(String.init)
        if let line = lines.first(where: { $0.lowercased().hasPrefix(key.lowercased()) }) {
            if let numStr = line.split(separator: ":").last?.replacingOccurrences(of: ".", with: "").trimmingCharacters(in: .whitespaces),
               let pages = Int64(numStr) { return pages }
        }
        return 0
    }
    let active = pages("Pages active")
    let wired = pages("Pages wired down") + pages("Pages wired")
    let compressed = pages("Pages occupied by compressor")
    let usedBytes = (active + wired + compressed) * Int64(pageSize)
    let usedMB = Int(usedBytes / 1_048_576)
    return (totalMB, usedMB)
}

func collectDisk() -> (totalMB: Int?, usedMB: Int?) {
    let fm = FileManager.default
    do {
        let attrs = try fm.attributesOfFileSystem(forPath: fm.homeDirectoryForCurrentUser.path)
        if let total = attrs[.systemSize] as? NSNumber, let free = attrs[.systemFreeSize] as? NSNumber {
            let totalMB = Int(truncating: total) / 1_048_576
            let freeMB = Int(truncating: free) / 1_048_576
            let usedMB = max(0, totalMB - freeMB)
            return (totalMB, usedMB)
        }
    } catch {}
    return (nil, nil)
}

func collectCPU() -> Double? {
    // Approximate: sum of %cpu across processes from ps; clamp 0..100
    let out = runShell("ps", ["-A", "-o", "%cpu="])
    let total = out.split(separator: "\n").compactMap { Double($0.trimmingCharacters(in: .whitespaces)) }.reduce(0, +)
    // Normalize rough upper bound; many-core can exceed 100, so cap to 100 for UI sanity
    return min(100.0, total)
}

func collectNet() -> (inBytes: Int?, outBytes: Int?) {
    // Sum for first non-loopback interface from netstat -ib
    let out = runShell("netstat", ["-ib"])
    let lines = out.split(separator: "\n").map(String.init)
    let rows = lines.dropFirst() // header
    for row in rows {
        let cols = row.split(separator: " ", omittingEmptySubsequences: true).map(String.init)
        if cols.count < 11 { continue }
        let iface = cols[0]
        if iface == "lo0" { continue }
        // Ibytes at col 6, Obytes at col 9 in many macOS versions; fallback tolerant
        let ib = Int(cols[safe: 6] ?? "") ?? Int(cols[safe: 7] ?? "")
        let ob = Int(cols[safe: 9] ?? "") ?? Int(cols[safe: 10] ?? "")
        if let ib = ib, let ob = ob { return (ib, ob) }
    }
    return (nil, nil)
}

extension Array {
    subscript(safe idx: Int) -> Element? { (0..<count).contains(idx) ? self[idx] : nil }
}

func collectMetrics() -> Metrics {
    let (tm, um) = collectMemory()
    let (td, ud) = collectDisk()
    let cpu = collectCPU()
    let (ib, ob) = collectNet()
    return Metrics(cpuUsage: cpu, memTotalMB: tm, memUsedMB: um, diskTotalMB: td, diskUsedMB: ud, netIn: ib, netOut: ob)
}

// MARK: - Registration & send
func registerDevice(config: Config) async {
    let regURL = config.baseURL.appendingPathComponent("devices/register")
    let body: [String: Any] = [
        "deviceid": config.deviceID,
        "device_name": hostName(),
        "device_type": "laptop",
        "os": osName(),
        "current_user": currentUser()
    ]
    do {
        let (status, _) = try await HTTP.postJSON(url: regURL, body: body)
        if status != 200 {
            fputs("[warn] register status=\(status)\n", stderr)
        }
    } catch {
        fputs("[warn] register failed: \(error)\n", stderr)
    }
}

func sendMetrics(config: Config, m: Metrics) async {
    let url = config.baseURL
        .appendingPathComponent("devices")
        .appendingPathComponent(config.deviceID)
        .appendingPathComponent("metrics")
    var body: [String: Any] = [:]
    if let v = m.cpuUsage { body["cpu_usage"] = v }
    if let v = m.memTotalMB { body["memory_total"] = v }
    if let v = m.memUsedMB { body["memory_used"] = v }
    if let v = m.diskTotalMB { body["disk_total"] = v }
    if let v = m.diskUsedMB { body["disk_used"] = v }
    if let v = m.netIn { body["net_bytes_in"] = v }
    if let v = m.netOut { body["net_bytes_out"] = v }
    do {
        let (status, _) = try await HTTP.postJSON(url: url, body: body)
        if status != 200 {
            fputs("[warn] metrics status=\(status)\n", stderr)
        }
    } catch {
        fputs("[warn] metrics failed: \(error)\n", stderr)
    }
}

// MARK: - Main loop
@main
struct Main {
    static func main() async {
        do {
            let cfg = try Config.load()
            print("DeviceAgent starting. deviceID=\(cfg.deviceID) base=\(cfg.baseURL) interval=\(cfg.intervalSeconds)s")

            // One-time registration (best-effort)
            await registerDevice(config: cfg)

            while true {
                let m = collectMetrics()
                await sendMetrics(config: cfg, m: m)
                try? await Task.sleep(nanoseconds: UInt64(cfg.intervalSeconds) * 1_000_000_000)
            }
        } catch {
            fputs("[error] \(error)\n", stderr)
            exit(1)
        }
    }
}
