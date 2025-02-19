import logging
from flask import Blueprint, render_template, current_app, redirect, request
from db.database import fetch_query

main = Blueprint('main', __name__)

@main.route('/')
def home():
    if current_app.spotify.is_authenticated:
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
        LIMIT 50;
        """
        recent_songs = fetch_query(query)
        history = [] # [{ 'cover_image': string, 'songs': Song[] }]
        if len(recent_songs) > 0:
            history.append({
                'cover_image': recent_songs[0]['cover_image'],
                'songs': [recent_songs[0]]
            })
            for i in range(1, len(recent_songs) - 1):
                hist_ind = len(history) - 1
                if recent_songs[i]['cover_image'] == history[hist_ind]['cover_image']:
                    history[hist_ind]['songs'].append(recent_songs[i])
                else:
                    history.append({
                        'cover_image': recent_songs[i]['cover_image'],
                        'songs': [recent_songs[i]]
                    })
        return render_template('recent.html', title="Recent", history=history)
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


# Single page routes (require valid id, will redirect otherwise)
@main.route('/song')
def song():
    track_id = request.args.get('id', '')
    if not track_id:
        return redirect('/')
    
    song = fetch_query(f"SELECT * FROM tracks WHERE id='{track_id}'")
    logging.info(song)
    return render_template('song.html', title=song[0]['name'], song=song)


@main.route('/album')
def album():
    album_id = request.args.get('id', '')
    if not album_id:
        return redirect('/')
    
    song = fetch_query(f"SELECT * FROM albums WHERE id='{album_id}'")
    logging.info(song)
    return render_template('album.html', title="Album", album=album)


@main.route('/artist')
def artist():
    artist_id = request.args.get('id', '')
    if not artist_id:
        return redirect('/')
    
    song = fetch_query(f"SELECT * FROM artists WHERE id='{artist_id}'")
    logging.info(song)
    return render_template('artist.html', title="Artist", artist=artist)
