import pathlib
import subprocess
from typing import Any

import requests

from dagster import op, Field, build_op_context

import re

@op(config_schema={
    "playlist_name": Field(str),
    "tempo": Field(float)
    }
)
def filter_and_modify_playlist(context) -> Any:
    
    playlist_name = context.op_config["playlist_name"]
    tempo = context.op_config["tempo"]

    path_to_filter = f"{pathlib.Path(__file__).parents[3]}/filter/filter"
    print(pathlib.Path(path_to_filter).resolve())
    command = [
        path_to_filter,
        "-playlist",
        playlist_name,
        "-tempo",
        str(tempo)
    ]
    
    url_regex = re.compile(r"^Please log in to Spotify by visiting the following page in your browser: (\S+)$")

    result = subprocess.Popen(command)
    print(result.stdout)
    match = re.search(url_regex, str(result.stdout))
    if match is not None:
        requests.get(match.group(0))
context = build_op_context(op_config={"playlist_name": "The Sound of Acid Techno", "tempo": 150.0})

filter_and_modify_playlist(context)
