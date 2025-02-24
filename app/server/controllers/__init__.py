from .SongController import SongController, SongControllerException
from .AlbumController import AlbumController, AlbumControllerException
from .ArtistController import ArtistController, ArtistControllerException

class Controller:
    song = SongController()
    album = AlbumController()
    artist = ArtistController()