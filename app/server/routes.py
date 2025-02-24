import logging
from flask import Blueprint, render_template, current_app, redirect, request

from server.controllers import Controller, SongControllerException, AlbumControllerException, ArtistControllerException

main = Blueprint('main', __name__)

@main.route('/')
def home():
    if current_app.spotify.is_authenticated:
        try:
            history = Controller.song.get_recents()
            return render_template('recent.html', title="Recent", history=history)
        except SongControllerException as e:
            logging.critical(f"HOMEPAGE ERROR {str(e)}")
    else:
        return redirect('/setup')        


@main.route('/setup')
def setup():
    code = request.args.get('code', '')
    if code:
        current_app.spotify.authenticate(refresh=False, code=code)
    
    if current_app.spotify.is_authenticated:
        return redirect('/')

    return redirect(current_app.spotify.get_auth_redirect_url())


@main.route('/most-played')
def most_played():
    return render_template('most_played.html', title="Most Played")


@main.route('/ratings')
def ratings():
    return render_template('ratings.html', title="Ratings")

@main.route('/all')
def all():
    return render_template('ratings.html', title="All")


# Single page routes (require valid id, will redirect otherwise)
@main.route('/song')
def song():
    try:
        song_id = request.args.get('id', '')
        song = Controller.song.get_one(song_id)
        return render_template(
            'item_page.html', 
            title=song['name'], 
            item_type="song", 
            image_url=song['cover_image'],
            item=song

        )
    except SongControllerException as e:
        logging.error(f"/song {str(e)}")
        return redirect('/')


@main.route('/album')
def album():
    try:
        album_id = request.args.get('id', '')
        album = Controller.album.get_one(album_id)
        return render_template(
            'item_page.html', 
            title=album['name'], 
            item_type="album", 
            image_url=album['cover_image'],
            item=album
        )
    except AlbumControllerException as e:
        logging.error(f"/album {str(e)}")
        return redirect('/')


@main.route('/artist')
def artist():
    try:
        artist_id = request.args.get('id', '')
        artist = Controller.artist.get_one(artist_id)
        return render_template(
            'item_page.html', 
            title=artist['name'], 
            item_type="artist", 
            image_url=artist['cover_image'],
            item=artist
        )
    except AlbumControllerException as e:
        logging.error(f"/artist {str(e)}")
        return redirect('/')