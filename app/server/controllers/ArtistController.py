from db.database import fetch_query

class ArtistControllerException(Exception):
    pass

class ArtistController:
    def get_one(self, artist_id):
        if not artist_id:
            raise ArtistControllerException(f"Invalid Artist id: {artist_id}")

        artist = fetch_query("SELECT * FROM artists WHERE id = ?", (artist_id,))
        if not artist or len(artist) != 1:
            raise ArtistControllerException(f"Error retrieving Artist by id: {artist_id}")

        return artist[0]
