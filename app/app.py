import logging
from flask import Flask, render_template, redirect, request
from apscheduler.schedulers.background import BackgroundScheduler

from filters.filters import datetimeformat
from database.database import initialize_database, fetch_query
from spotify.spotify import SpotifyController

logging.basicConfig(
    level=logging.INFO,
    format="%(message)s",
    handlers=[logging.StreamHandler()]
)

initialize_database()

app = Flask(__name__)
spotify = SpotifyController()

scheduler = BackgroundScheduler()
scheduler.add_job(spotify.get_now_playing, 'interval', seconds=20)
scheduler.start()

@app.template_filter('datetimeformat')
def datetimeformatfilter(value, f='%b %d, %Y'):
    return datetimeformat(value, f)

@app.route('/')
def home():
    if spotify.is_authenticated:
        query = """
        SELECT 
            p.timestamp, 
            t.name AS track_name, 
            a.name AS artist_name, 
            alb.cover_image, 
            t.id as track_id,
            alb.id as album_id,
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
        return render_template('index.html', history=history)
    else:
        return redirect('/setup')        


@app.route('/setup')
def setup():
    code = request.args.get('code', '')
    if code:
        spotify.authenticate(code)
    
    if spotify.is_authenticated:
        return redirect('/')

    return redirect(spotify.get_auth_redirect_url())


if __name__ == '__main__':
    app.run(host='0.0.0.0', port=5000, debug=True)
