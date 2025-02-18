import os
import sqlite3
import logging
from datetime import datetime

DB_FILE = "db/db/database.db"
SCHEMA_FILE = "db/schema.sql"

def initialize_database():
    logging.info("Initializing database.")
    conn = sqlite3.connect(DB_FILE)
    cursor = conn.cursor()

    try:
        with open(SCHEMA_FILE, "r") as f:
            schema = f.read()
        cursor.executescript(schema)
        conn.commit()
    except sqlite3.Error as e:
        logging.error(f"SQLite error during initialization: {str(e)}")
    finally:
        conn.close()


def fetch_query(query, params=()):
    conn = sqlite3.connect(DB_FILE)
    conn.row_factory = sqlite3.Row 
    cursor = conn.cursor()

    try:
        cursor.execute(query, params)
        rows = cursor.fetchall()
        results = [dict(row) for row in rows]
        return results
    except sqlite3.Error as e:
        logging.error(f"fetch_query error: ${e}")
    finally:
        conn.close()
    
    return []
    

def execute_query(query, params=()):
    conn = sqlite3.connect(DB_FILE)
    cursor = conn.cursor()

    try:
        cursor.execute(query, params)
        conn.commit()
        return True
    except sqlite3.Error as e:
        logging.error(f"execute_query error: {e}")
        conn.rollback()
        return False
    finally:
        conn.close()


def get_spotify_creds():
    creds = fetch_query("SELECT * FROM spotify_auth")
    if (len(creds) != 1):
        logging.error("Invalid spotify credentials. Database failed to initialize")
        return None
    return {
        "access_token": creds[0]['access_token'],
        "refresh_token": creds[0]['refresh_token'],
        "expiry": int(creds[0]['expiry']),
    }


def set_spotify_creds(access, refresh, expiry):
    query = """
        INSERT INTO spotify_auth (id, access_token, refresh_token, expiry)
        VALUES (1, ?, ?, ?)
        ON CONFLICT(id) DO UPDATE 
        SET access_token = excluded.access_token,
            refresh_token = excluded.refresh_token,
            expiry = excluded.expiry;
    """
    return execute_query(query, (access, refresh, expiry))


def insert_playback_data(data):
    """Inserts Spotify playback data into the SQLite database."""

    if not data.get('item'):
        print("No song currently playing.")
        return
    
    track_info = data['item']
    album_info = track_info['album']
    artist_info = track_info['artists'][0]  # Get first artist
    device_info = data.get('device', {})

    # Extract relevant data
    album_id = album_info['id']
    album_name = album_info['name']
    album_release_date = album_info['release_date']
    album_url = album_info['external_urls']['spotify']
    album_type = album_info['album_type']
    album_cover = album_info['images'][0]['url'] if album_info.get('images') else None

    artist_id = artist_info['id']
    artist_name = artist_info['name']
    artist_url = artist_info['external_urls']['spotify']

    track_id = track_info['id']
    track_name = track_info['name']
    track_duration = track_info['duration_ms']
    track_explicit = track_info['explicit']
    track_popularity = track_info['popularity']
    track_url = track_info['external_urls']['spotify']
    track_number = track_info['track_number']

    device_id = device_info.get('id', 'unknown')
    device_name = device_info.get('name', 'Unknown Device')
    volume_percent = device_info.get('volume_percent', 0)

    progress_ms = data.get('progress_ms', 0)
    is_playing = data.get('is_playing', False)
    repeat_state = data.get('repeat_state', 'off')
    shuffle_state = data.get('shuffle_state', False)
    
    timestamp = datetime.now()

    # Connect to SQLite database
    conn = sqlite3.connect(DB_FILE)
    cursor = conn.cursor()

    try:
        # Insert Album (if not exists)
        cursor.execute("""
            INSERT INTO albums (id, name, release_date, spotify_url, album_type, cover_image)
            VALUES (?, ?, ?, ?, ?, ?)
            ON CONFLICT(id) DO NOTHING;
        """, (album_id, album_name, album_release_date, album_url, album_type, album_cover))

        # Insert Artist (if not exists)
        cursor.execute("""
            INSERT INTO artists (id, name, spotify_url)
            VALUES (?, ?, ?)
            ON CONFLICT(id) DO NOTHING;
        """, (artist_id, artist_name, artist_url))

        # Insert Track (if not exists)
        cursor.execute("""
            INSERT INTO tracks (id, name, duration_ms, explicit, popularity, spotify_url, track_number, album_id)
            VALUES (?, ?, ?, ?, ?, ?, ?, ?)
            ON CONFLICT(id) DO NOTHING;
        """, (track_id, track_name, track_duration, track_explicit, track_popularity, track_url, track_number, album_id))

        # Insert Track-Artist Relationship (if not exists)
        cursor.execute("""
            INSERT INTO track_artists (track_id, artist_id)
            VALUES (?, ?)
            ON CONFLICT(track_id, artist_id) DO NOTHING;
        """, (track_id, artist_id))

        # Insert Play History
        cursor.execute("""
            INSERT INTO play_history (timestamp, progress_ms, is_playing, device_id, device_name, volume_percent, repeat_state, shuffle_state, track_id)
            VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?);
        """, (timestamp, progress_ms, is_playing, device_id, device_name, volume_percent, repeat_state, shuffle_state, track_id))

        # Commit changes
        conn.commit()
        print("Data inserted successfully.")

    except sqlite3.Error as e:
        print("SQLite error:", e)
        conn.rollback()
    finally:
        logging.info(f"logged {track_name}")
        conn.close()
