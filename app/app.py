import logging
from flask import Flask, render_template, redirect, request
from apscheduler.schedulers.background import BackgroundScheduler

from database.database import initialize_database, fetch_query
from spotify.spotify import SpotifyController

logging.basicConfig(level=logging.INFO)

initialize_database()

app = Flask(__name__)
spotify = SpotifyController()

scheduler = BackgroundScheduler()
scheduler.add_job(spotify.get_now_playing, 'interval', seconds=20)
scheduler.start()

@app.route('/')
def home():
    if spotify.is_authenticated:
        query = """
        SELECT p.timestamp, t.name AS track_name, a.name AS artist_name
        FROM play_history p
        JOIN tracks t ON p.track_id = t.id
        JOIN track_artists ta ON t.id = ta.track_id
        JOIN artists a ON ta.artist_id = a.id
        ORDER BY p.timestamp DESC
        LIMIT 5;
        """
        recent_songs = fetch_query(query)
        return render_template('index.html', recent_songs=recent_songs)
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
