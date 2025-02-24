from datetime import datetime, timezone

def datetimeformat(value, f='%Y-%m-%d %H:%M:%S'):
    dt = get_dt(value)
    if f == "recent":
        return recent_datetimeformat(dt)
    
    return dt.strftime(f)


def get_dt(value):
    if isinstance(value, (int, float)):  
        return datetime.fromtimestamp(value, timezone.utc)
    try:
        return datetime.fromisoformat(value)
    except ValueError:
        return value


def recent_datetimeformat(dt):
    now = datetime.now()
    if dt.tzinfo is not None: 
        now = now.replace(tzinfo=timezone.utc)
    
    diff = now - dt

    seconds = diff.total_seconds()

    def add_s(v):
        return "s" if v != 1 else ""

    val = 0
    if seconds < 10:
        return "Now"
    elif seconds < 60:
        val = int(seconds)
        return f"{val} sec{add_s(val)}"
    elif seconds < 3600:
        val = int(seconds // 60)
        return f"{val} min{add_s(val)}"
    elif seconds < 86400:
        val = int(seconds // 3600)
        return f"{val} hr{add_s(val)}"
    elif seconds < 604800:  # 7 days
        val = int(seconds // 86400)
        return f"{val} day{add_s(val)}"
    else:
        return dt.strftime('%b %d, %Y') 
    
