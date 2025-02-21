from flask import Blueprint, jsonify, request

api = Blueprint('api', __name__)

@api.route('/top', methods=['GET'])
def get_top():
    """
    Get the played items (songs, albums or artists) based on type.

    Args:
        category (str, optional): The type of ranking to retrieve. Defaults to 'songs'.
            - Allowed values: `songs`, `albums`, `artists`
        limit (int, optional): Number of top items to get
            - must be between 1 and 1000 (until performance tests)
            - defaults to 100
        
        If an invalid category is provided, a 400 error is returned.

    Returns:
        JSON: A list of top 100 items in the selected category with the following format:
        ```json
        [
            {
                "id": Item.id
                "name": "Song Title" | "Artist Name" | "Album Name",
                "artist": "Artist Name",
                "cover_image": "Album Art Url",
                "plays": 12345
            }
        ]
        ```
    """
    valid_categories = ["songs", "artists", "albums"]
    category = request.args.get('category', 'songs')
    if category not in valid_categories:
        return jsonify({'error': f'Invalid category \'{category}\''})
    limit = int(request.args.get('limit', 100))
    if limit < 1 or limit > 1000:
        return jsonify({'error': f'Invalid limit {limit}. Must be [1,1000]'})
    

    if category == 'songs':
        results = []
        # data = Song.query.order_by(Song.play_count.desc()).limit(100).all()
        # results = [{'title': s.title, 'artist': s.artist, 'plays': s.play_count} for s in data]
    else:
        return jsonify({'error': 'Invalid category'}), 400

    return jsonify(results)
