"""Optional one-time landmark extraction; rebuilding the GLB uses cached JSON.

Requires mediapipe==0.10.35 in addition to requirements.txt, with its standard
dependencies installed, and the official face_landmarker.task model.
"""
import argparse
import json
from pathlib import Path

import mediapipe as mp

root = Path(__file__).resolve().parents[2]
parser = argparse.ArgumentParser(description=__doc__)
parser.add_argument("--model", type=Path, default=root / ".tools/art/face_landmarker.task")
parser.add_argument("--source", type=Path, default=root / "assets/source/ivanov/face-reference.png")
parser.add_argument("--output", type=Path, default=root / "assets/source/ivanov/face-landmarks.json")
args = parser.parse_args()
options = mp.tasks.vision.FaceLandmarkerOptions(
    base_options=mp.tasks.BaseOptions(model_asset_path=str(args.model)),
)
with mp.tasks.vision.FaceLandmarker.create_from_options(options) as detector:
    result = detector.detect(mp.Image.create_from_file(str(args.source)))
if len(result.face_landmarks) != 1:
    raise RuntimeError("Expected exactly one visible face")
points = [[landmark.x, landmark.y, landmark.z] for landmark in result.face_landmarks[0]]
args.output.parent.mkdir(parents=True, exist_ok=True)
args.output.write_text(json.dumps(points), encoding="utf-8")
print(f"Saved {len(points)} landmarks to {args.output}")
