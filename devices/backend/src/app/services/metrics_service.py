MAX_PERCENT = 100


class MetricsService:
    """Stub MetricsService for tests."""

    def validate_cpu_usage(self, v):
        return 0 <= v <= MAX_PERCENT

    def validate_memory_usage(self, used, total):
        return used >= 0 and total is not None and used <= total

    def validate_disk_usage(self, used, total):
        return used >= 0 and used <= total

    def calculate_average_metrics(self, data):
        if not data:
            return {"avg_cpu": 0, "avg_memory": 0}
        avg_cpu = sum(d.get("cpu_usage", 0) for d in data) / len(data)
        avg_mem = sum(d.get("memory_used", 0) for d in data) / len(data)
        return {"avg_cpu": avg_cpu, "avg_memory": avg_mem}
