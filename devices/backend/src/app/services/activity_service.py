class ActivityService:
    """Stub ActivityService for tests."""
    def validate_activity_type(self, t):
        return t in {"app_launch","app_close","file_access","web_visit","idle"}
    def validate_duration(self, d):
        return d >= 0
    def analyze_activity_patterns(self, activities):
        from collections import Counter
        apps = [a.get("app") for a in activities if a.get("app")]
        cnt = Counter(apps)
        most = cnt.most_common(1)[0][0] if cnt else None
        return {"most_used_app": most, "total_activities": len(activities), "unique_apps": len(cnt)}
