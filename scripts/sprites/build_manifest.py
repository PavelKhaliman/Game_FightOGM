"""Read chroma-key sheets and write runtime frame rectangles; never alter PNGs.

Run with .tools/art/venv/Scripts/python.exe scripts/sprites/build_manifest.py.
Connected components preserve extended fists that cross nominal cell borders.
"""
import json
from pathlib import Path

import numpy as np
from PIL import Image

ROOT = Path(__file__).resolve().parents[2]
SPRITES = ROOT / 'assets/sprites/ivanov'


def components(mask, count=6, rows=2, details=False):
    parents, spans, previous = [], [], []

    def root(i):
        while parents[i] != i:
            parents[i] = parents[parents[i]]
            i = parents[i]
        return i

    for y, row in enumerate(mask):
        edges = np.flatnonzero(np.diff(np.r_[False, row, False]))
        current = []
        for x0, x1 in zip(edges[::2], edges[1::2]):
            i = len(parents)
            parents.append(i)
            spans.append((int(x0), y, int(x1)))
            for a, b, j in previous:
                if a > x1:
                    break
                if b >= x0:
                    parents[root(j)] = root(i)
            current.append((x0, x1, i))
        previous = current
    boxes = {}
    for i, (x0, y, x1) in enumerate(spans):
        k = root(i)
        if k not in boxes:
            boxes[k] = [x0, y, x1, y + 1, x1 - x0]
        else:
            b = boxes[k]
            b[0], b[1] = min(b[0], x0), min(b[1], y)
            b[2], b[3] = max(b[2], x1), max(b[3], y + 1)
            b[4] += x1 - x0
    big = sorted(boxes, key=lambda k: -boxes[k][4])[:count]
    assert len(big) == count and min(boxes[k][4] for k in big) > 3000, big
    big.sort(key=lambda k: (int((boxes[k][1] + boxes[k][3]) / 2 / (mask.shape[0] / rows)), boxes[k][0]))
    if details:
        selected = {k: [] for k in big}
        for i, span in enumerate(spans):
            k = root(i)
            if k in selected:
                selected[k].append(span)
        return [{'bounds': boxes[k], 'spans': selected[k]} for k in big]
    return [boxes[k] for k in big]


def main():
    heights = {
        'stances': [2, 2, 2, 1.96, 1.40, 1.52],
        'walk': [2, 2, 2, 2, 2, 2],
        'punches': [2, 2, 2, 2, 1.96, 2],
        'kicks': [2, 1.94, 2, 1.39, 1.31, 1.46],
        'reactions': [1.94, 1.66, .55, 1.35, 1.85, 2.30],
        'specials': [2, 2, 1.92, 1.90, 2, 2.05],
    }
    frames = {}
    for sheet, world_heights in heights.items():
        path = SPRITES / (sheet + '.png')
        if not path.exists():
            print('Pending:', path.name)
            continue
        pixels = np.asarray(Image.open(path).convert('RGB')).astype(np.int16)
        red, green, blue = pixels[:, :, 0], pixels[:, :, 1], pixels[:, :, 2]
        mask = green - np.maximum(red, blue) < 75
        for index, (left, top, right, bottom, area) in enumerate(components(mask)):
            x, y = max(0, left - 2), max(0, top - 2)
            w, h = min(pixels.shape[1], right + 2) - x, min(pixels.shape[0], bottom + 2) - y
            # Use the supporting shoe, excluding hands and the extended attacking leg.
            feet = mask[max(top, bottom - int((bottom - top) * .06)):bottom, left:right]
            feet_x = np.flatnonzero(feet.any(axis=0))
            px = float(left + (feet_x[0] + feet_x[-1]) / 2 - x)
            if sheet == 'reactions' and index == 2:
                px = w / 2
            if sheet == 'kicks' and index in (3, 4):
                # The extended attacking foot is not the ground anchor of a sweep.
                px = {3: 60.0, 4: 95.0}[index]
            frame = {'sheet': path.name, 'rect': [x, y, w, h],
                     'pivot': [px, float(bottom - y)], 'world_height': world_heights[index]}
            frames[f'{sheet}_{index}'] = frame
            print(f'{sheet}_{index}: {frame["rect"]}, foot {frame["pivot"]}, area {area}')

    clips = {}


    def clip(name, sheet, indices, fps=8, loop=False, attack=False):
        clips[name] = {'frames': [f'{sheet}_{i}' for i in indices], 'fps': fps,
                       'loop': loop, 'attack': attack}


    clip('idle', 'stances', [0, 1, 0, 2], 4, True)
    clip('walk', 'walk', range(6), 9, True)
    clip('block', 'stances', [3], 6, True)
    clip('crouch', 'stances', [4])
    clip('jump', 'stances', [5])
    clip('punch_light', 'punches', [0, 1, 2], attack=True)
    clip('punch_heavy', 'punches', [3, 4, 5], attack=True)
    clip('kick_high', 'kicks', [0, 1, 2], attack=True)
    clip('kick_low', 'kicks', [3, 4, 5], attack=True)
    clip('hit_react', 'reactions', [0])
    clip('knockdown', 'reactions', [1, 2, 2, 2])
    clip('get_up', 'reactions', [2, 3, 4])
    clip('victory', 'reactions', [5], 3, True)
    clip('grab', 'specials', [2, 3, 2], attack=True)
    clip('bad_hearing', 'specials', [0, 1])
    clip('production_invisible', 'specials', [4, 5])
    clip('omni_transform', 'specials', [4, 5])
    clip('omni_rush', 'punches', [3, 4, 5], attack=True)
    if len(frames) == 36:
        (SPRITES / 'animations.json').write_text(json.dumps({'frames': frames, 'clips': clips}, indent=2) + '\n', encoding='utf-8')
        print('Wrote 36 frames / 18 clips.')


if __name__ == "__main__":
    main()
