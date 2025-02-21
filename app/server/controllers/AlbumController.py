from db.database import fetch_query

class AlbumControllerException(Exception):
    pass

class AlbumController:
    def get_one(self, album_id):
        if not album_id:
            raise AlbumControllerException(f"Invalid Album id: {album_id}")

        album = fetch_query("SELECT * FROM albums WHERE id = ?", (album_id,))
        if not album or len(album) != 1:
            raise AlbumControllerException(f"Error retrieving Album by id: {album_id}")

        return album[0]
