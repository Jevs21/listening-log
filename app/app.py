import logging
from flask import Flask, render_template, redirect, request
from apscheduler.schedulers.background import BackgroundScheduler

from filters.filters import datetimeformat
from database.database import initialize_database
from spotify.spotify import SpotifyController

logging.basicConfig(
    level=logging.INFO,
    format="%(message)s",
    handlers=[logging.StreamHandler()]
)

initialize_database()

from routes import main as main_blueprint
from api import api as api_blueprint

app = Flask(__name__)
app.register_blueprint(main_blueprint)
app.register_blueprint(api_blueprint, url_prefix='/api')
app.spotify = SpotifyController()

scheduler = BackgroundScheduler()
scheduler.add_job(app.spotify.get_now_playing, 'interval', seconds=20)
scheduler.start()

@app.template_filter('datetimeformat')
def datetimeformatfilter(value, f='%b %d, %Y'):
    return datetimeformat(value, f)


if __name__ == '__main__':
    app.run(host='0.0.0.0', port=5000, debug=True)
