"""Read generated sheets; derive JSON only, leaving every PNG unchanged."""
import json
from pathlib import Path
import numpy as np
from PIL import Image
from build_manifest import components

ROOT = Path(__file__).resolve().parents[2]
PEOPLE = json.loads((Path(__file__).parent / 'roster_people.json').read_text(encoding='utf-8'))
HEIGHTS = {
    'movement': [2, 2, 2, 1.96, 1.4, 1.52, 2, 2, 2, 2, 2, 2],
    'attacks': [2, 2, 2, 2, 1.96, 2, 2, 1.94, 2, 1.39, 1.31, 1.46],
    'actions': [1.94, 1.66, .55, 1.35, 1.85, 2.30, 2, 1.96, 2, 1.94, 2.05, 2.08],
}
SPECIALS = {
    'gagloev': ('skewer_control', 'block', 'shashlik'),
    'kiselik': ('sueta_dash', 'yes_boss', 'everything_at_once'),
    'novatskiy': ('screw_throw', 'snap_back', 'screw_batch'),
    'zhirnov': ('ai_agent', 'autopilot', 'do_it_for_me'),
    'elkhimov': ('wrench_32', 'repair_armor', 'overhaul'),
    'kalachev': ('water_burst', 'was_not_here', 'break_return'),
    'fedoseev': ('sarcasm_wave', 'slow_clap', 'genius_plan'),
    'khaliman': ('plc_pulse', 'gym_mode', 'asutp'),
    'shuev': ('karate_flurry', 'stance_counter', 'endless_shift'),
}


def build(person):
    folder = ROOT / 'assets/sprites' / person['id']
    frames = {}
    for sheet, heights in HEIGHTS.items():
        path = folder / (sheet + '.png')
        if not path.exists():
            print(person['id'], 'pending', path.name)
            continue
        pixels = np.asarray(Image.open(path).convert('RGB')).astype(np.int16)
        r, g, b = pixels[:, :, 0], pixels[:, :, 1], pixels[:, :, 2]
        mask = g - np.maximum(r, b) < 75
        figures = components(mask, count=12, rows=3, details=True)
        boxes = [figure['bounds'] for figure in figures]
        row_counts = [sum(int((b[1]+b[3])/2/(mask.shape[0]/3)) == row for b in boxes) for row in range(3)]
        assert row_counts == [4, 4, 4], (path, row_counts)
        for index, (left, top, right, bottom, area) in enumerate(boxes):
            x, y = max(0, left-2), max(0, top-2)
            w, h = min(pixels.shape[1], right+2)-x, min(pixels.shape[0], bottom+2)-y
            # Atlas cells may overlap at their corners. Split the source rectangle
            # into a few disjoint draw regions, retaining only this figure's extent
            # in each band. This is rendering metadata; source pixels stay intact.
            cuts = {y, y+h}
            for other, box in enumerate(boxes):
                if other == index:
                    continue
                bx, by, br, bb, _ = box
                if min(x+w, br+2) > max(x, bx-2) and min(y+h, bb+2) > max(y, by-2):
                    cuts.update([max(y, by-2), min(y+h, bb+2)])
            regions = []
            cuts = sorted(cuts)
            if len(cuts) > 2:
                for y0, y1 in zip(cuts, cuts[1:]):
                    spans = [s for s in figures[index]['spans'] if y0-2 <= s[1] < y1+2]
                    if not spans:
                        continue
                    intervals = sorted((max(x, s[0]-2), min(x+w, s[2]+2)) for s in spans)
                    merged = []
                    for x0, x1 in intervals:
                        if merged and x0 <= merged[-1][1]:
                            merged[-1][1] = max(merged[-1][1], x1)
                        else:
                            merged.append([x0, x1])
                    for x0, x1 in merged:
                        regions.append([x0-x, y0-y, x1-x0, y1-y0])
            feet = mask[max(top, bottom-int((bottom-top)*.06)):bottom, left:right]
            columns = np.flatnonzero(feet.any(axis=0))
            px = float(left+(columns[0]+columns[-1])/2-x)
            if sheet == 'attacks' and index in (9, 10):
                # Anchor to the supporting foot, not the extended sweep foot.
                split = np.flatnonzero(np.diff(columns) > 8)
                last = split[0] if len(split) else len(columns)-1
                px = float(left+(columns[0]+columns[last])/2-x)
            if sheet == 'actions' and index == 2:
                px = w/2
            if sheet == 'movement' and index == 5:
                px = w/2
            world_height = heights[index]
            if sheet == 'actions' and index >= 6:
                # Keep the same anatomical scale as the upright get-up frame;
                # crouched attacks must lower the head instead of enlarging it.
                reference_height = boxes[4][3]-boxes[4][1]+4
                world_height = round(1.95*h/reference_height, 4)
            frame = {
                'sheet': path.name, 'rect': [x, y, w, h],
                'pivot': [px, float(bottom-y)], 'world_height': world_height
            }
            if regions:
                frame['regions'] = regions
            frames[f'{sheet}_{index:02d}'] = frame
        print(person['id'], sheet, len(boxes), 'frames', tuple(pixels.shape[:2][::-1]))
    if len(frames) != 36:
        return
    clips = {}

    def clip(name, sheet, indices, fps=8, loop=False, attack=False):
        clips[name] = {'frames': [f'{sheet}_{i:02d}' for i in indices],
                       'fps': fps, 'loop': loop, 'attack': attack}

    clip('idle', 'movement', [0, 1, 0, 2], 4, True)
    clip('walk', 'movement', range(6, 12), 9, True)
    clip('block', 'movement', [3], 6, True)
    clip('crouch', 'movement', [4])
    clip('jump', 'movement', [5])
    for name, start in [('punch_light', 0), ('punch_heavy', 3), ('kick_high', 6), ('kick_low', 9)]:
        clip(name, 'attacks', range(start, start+3), attack=True)
    clip('hit_react', 'actions', [0])
    clip('knockdown', 'actions', [1, 2, 2, 2])
    clip('get_up', 'actions', [2, 3, 4])
    clip('victory', 'actions', [5], 3, True)
    clips['grab'] = {'frames': ['movement_00', 'actions_10', 'movement_00'],
                     'fps': 8, 'loop': False, 'attack': True}
    special, secondary, ultimate = SPECIALS[person['id']]
    clip(special, 'actions', [6, 7, 8], attack=True)
    if secondary != 'block':
        clip(secondary, 'actions', [9])
    clip(ultimate, 'actions', [10, 11, 8], attack=True)
    if person['id'] == 'shuev':
        clip(ultimate, 'actions', [10, 11])
    if person['id'] == 'khaliman':
        clip('standby', 'movement', [0, 1], 2, True)
    (folder/'animations.json').write_text(json.dumps({'frames': frames, 'clips': clips}, indent=2)+'\n', encoding='utf-8')
    print('READY:', person['id'], len(frames), 'frames /', len(clips), 'clips')


if __name__ == '__main__':
    for person in PEOPLE:
        build(person)
