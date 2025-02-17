import os
import time
import requests
import logging
from urllib.parse import urlencode
from db.database import insert_playback_data, get_spotify_creds, set_spotify_creds

logging.basicConfig(level=logging.INFO)

CLIENT_ID = os.getenv("CLIENT_ID")
CLIENT_SECRET = os.getenv("CLIENT_SECRET")
REDIRECT_URI = "http://localhost/setup" # TODO: Change this
SCOPE = "user-read-currently-playing"

class SpotifyController:
    def __init__(self):
        self.is_authenticated = False
        self.auth_code    = ""
        self.access_token = ""
        self.refresh_tok  = ""
        self.expiry       = 0

        auth_data = get_spotify_creds()
        if auth_data:
            self.is_authenticated = True # Will refresh if it needs it
            self.access_token     = auth_data['access_token']
            self.refresh_token    = auth_data['refresh_token']
            self.expiry           = auth_data['expiry']

    def __repr__(self):
        return f"Spotify(auth={self.is_authenticated},expiry={self.expiry})"
 
    def get_auth_redirect_url(self):
        params = { 
            "client_id": CLIENT_ID,
            "response_type": "code",
            "redirect_uri": REDIRECT_URI,
            "scope": SCOPE
        }
        return f"https://accounts.spotify.com/authorize?{urlencode(params)}"

    def set_auth_data()
    def has_token_expired(self):
        return time.time() > self.expiry

    def authenticate(self, refresh = False, code = None):
        if code:
            self.auth_code = code
        
        headers = { "Content-Type": "application/x-www-form-urlencoded" }
        payload = {}
        if refresh and self.refresh_tok:
            payload = {
                "grant_type": "refresh_token",
                "refresh_token": self.refresh_tok,
                "client_id": CLIENT_ID
            }
        else:
            payload = { 
                "grant_type": "authorization_code", 
                "code": self.auth_code,
                "redirect_uri": REDIRECT_URI, 
                "client_id": CLIENT_ID, 
                "client_secret": CLIENT_SECRET 
            }
        
        logging.info(f"Using {payload}")
        response = requests.post("https://accounts.spotify.com/api/token", headers=headers, data=payload)
        # try:
        response.raise_for_status()
        data = response.json()

        self.access_token = data['access_token']
        self.refresh_tok = data['refresh_token']
        self.expiry = time.time() + data['expires_in']
        self.is_authenticated = True
        # except e:
        #     self.is_authenticated = False
    

    def get_now_playing(self):
        if not self.is_authenticated:
            logging.info("Not authenticated: Visit http://localhost/ to authenticate")
            return
        if self.has_token_expired():
            logging.info("Reauthenticating")
            self.authenticate(refresh=True)
            return

        headers = { "Authorization": f"Bearer {self.access_token}"}
        response = requests.get("https://api.spotify.com/v1/me/player/currently-playing", headers=headers)
        if response.status_code == 204:
            logging.info("No song playing")
            return
        
        response.raise_for_status()
        data = response.json()
        try:
            insert_playback_data(data)
        except Exception as e:
            logging.error(e)

        

