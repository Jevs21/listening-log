from db.database import fetch_query

class SongControllerException(Exception):
    pass

class SongController:
    def get_one(self, song_id):
        if not song_id:
            raise SongControllerException(f"Invalid song id: {song_id}")
        query = """
            SELECT
                t.name as name,
                a.name as artist_name,
                alb.cover_image
            FROM tracks t
            JOIN track_artists ta ON t.id = ta.track_id
            JOIN artists a ON ta.artist_id = a.id
            JOIN albums alb ON t.album_id = alb.id
            WHERE t.id = ?
        """
        song = fetch_query(query, (song_id,))
        if not song or len(song) != 1:
            raise SongControllerException(f"Error retrieving song by id: {song_id}")

        return song[0]


    def get_recents(self, limit=50):
        query = """
            SELECT 
                p.timestamp, 
                t.name AS track_name, 
                a.name AS artist_name, 
                alb.cover_image, 
                t.id as track_id,
                alb.id as album_id,
                alb.name as album_name,
                a.id as artist_id
            FROM play_history p
            JOIN tracks t ON p.track_id = t.id
            JOIN track_artists ta ON t.id = ta.track_id
            JOIN artists a ON ta.artist_id = a.id
            JOIN albums alb ON t.album_id = alb.id
            WHERE p.timestamp = (
                SELECT MAX(p2.timestamp)
                FROM play_history p2
                JOIN tracks t2 ON p2.track_id = t2.id
                WHERE t2.name = t.name
            )
            ORDER BY p.timestamp DESC
            LIMIT ?;
        """
        recent_songs = fetch_query(query, (limit,))
        recent = [] # [{ 'cover_image': string, 'songs': Song[] }]
        if len(recent_songs) > 0:
            recent.append({
                'cover_image': recent_songs[0]['cover_image'],
                'songs': [recent_songs[0]]
            })
            for i in range(1, len(recent_songs) - 1):
                hist_ind = len(recent) - 1
                if recent_songs[i]['cover_image'] == recent[hist_ind]['cover_image']:
                    recent[hist_ind]['songs'].append(recent_songs[i])
                else:
                    recent.append({
                        'cover_image': recent_songs[i]['cover_image'],
                        'songs': [recent_songs[i]]
                    })
        
        return recent
