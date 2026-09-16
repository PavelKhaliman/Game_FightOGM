"""Build a textured, skinned Ivanov GLB without changing the supplied rig/clips.

The face uses measured MediaPipe landmarks and its Apache-2.0 face topology.
All remaining surfaces are authored here in the original model's Z-up bind pose.
NumPy and Pillow rebuild it; the measured landmarks and PNGs are checked in.
"""
from __future__ import annotations

import argparse
import copy
import hashlib
import json
import math
from pathlib import Path
import struct

import numpy as np
from PIL import Image

ROOT = Path(__file__).resolve().parents[2]
SOURCE = ROOT / "assets/source/ivanov"
TAU = math.tau
SKIN = (.69, .46, .36, 1)


def unit(v):
    v = np.asarray(v, dtype=float)
    return v / max(1e-12, np.linalg.norm(v))


def blend(a, b, t):
    t = float(np.clip(t, 0, 1))
    return {a: 1 - t, b: t}


def torso_weights(z):
    if z < 1.13:
        return blend(1, 2, (z - .99) / .14)
    return blend(2, 3, (z - 1.13) / .19)


def atlas(quadrant, u, v):
    # Insets keep bilinear filtering away from neighbouring atlas materials.
    return ((quadrant % 2) * .5 + .008 + u * .484,
            (quadrant // 2) * .5 + .008 + v * .484)


class Surface:
    def __init__(self, name, material):
        self.name, self.material = name, material
        self.p, self.uv, self.colors, self.joints, self.weights, self.faces = [], [], [], [], [], []

    def vertex(self, p, uv=(0, 0), color=(1, 1, 1, 1), weights=None):
        if self.material == 3:
            # Reuse a clean cheek region as fine skin albedo on the neck/hands;
            # the PNG is unchanged, only the mesh UVs and colour factors vary.
            uv = (.616 + .052 * (math.sin(p[0] * 9) + 1) / 2,
                  .484 + .080 * (math.sin(p[2] * 7) + 1) / 2)
            color = tuple(float(np.clip(c / base, 0, 1)) for c, base in zip(color[:3], (.77, .50, .40))) + (color[3],)
        weights = weights or {3: 1}
        pairs = sorted(((j, w) for j, w in weights.items() if w > 1e-8), key=lambda x: -x[1])[:4]
        total = sum(w for _, w in pairs)
        self.p.append(list(p))
        self.uv.append(uv)
        self.colors.append(color)
        self.joints.append([j for j, _ in pairs] + [0] * (4 - len(pairs)))
        self.weights.append([w / total for _, w in pairs] + [0] * (4 - len(pairs)))
        return len(self.p) - 1

    def tri(self, a, b, c, outward=None):
        if outward is not None:
            pa, pb, pc = np.asarray([self.p[k] for k in (a, b, c)])
            if np.dot(np.cross(pb - pa, pc - pa), outward) < 0:
                b, c = c, b
        self.faces.append([a, b, c])

    def quad(self, a, b, c, d, outward=None):
        self.tri(a, b, c, outward)
        self.tri(a, c, d, outward)

    def normals(self):
        p = np.asarray(self.p)
        f = np.asarray(self.faces)
        n = np.zeros_like(p)
        fn = np.cross(p[f[:, 1]] - p[f[:, 0]], p[f[:, 2]] - p[f[:, 0]])
        for i in range(3):
            np.add.at(n, f[:, i], fn)
        # Weld coincident positions for smooth UV seams, without welding geometry.
        groups = {}
        for i, pos in enumerate(p):
            groups.setdefault(tuple(np.round(pos, 7)), []).append(i)
        for ids in groups.values():
            if len(ids) > 1:
                n[ids] = n[ids].sum(axis=0)
        length = np.linalg.norm(n, axis=1)
        n[length < 1e-12] = [0, 0, 1]
        return n / np.maximum(np.linalg.norm(n, axis=1)[:, None], 1e-12)


def tube(surf, centers, radii, weights, quadrant=None, color=(1, 1, 1, 1), segments=32, folds=0, caps=True, repeat=None, cutout=None):
    centers = np.asarray(centers)
    source_centers, source_radii, source_weights = centers.copy(), np.asarray(radii), weights
    repeat = repeat or ((3, 2) if quadrant == 0 and len(centers) > 8 else (3, 5) if quadrant == 1 and len(centers) > 8 else (1, 1))
    ts = sorted(set(np.linspace(0, 1, len(centers)).tolist() + np.linspace(0, 1, repeat[1] + 1).tolist()))
    original_ts = np.linspace(0, 1, len(centers))
    centers = np.column_stack([np.interp(ts, original_ts, source_centers[:, i]) for i in range(3)])
    radii = np.column_stack([np.interp(ts, original_ts, source_radii[:, i]) for i in range(2)])
    segments = math.ceil(segments / repeat[0]) * repeat[0]
    if callable(source_weights):
        weights = lambda j, p: source_weights(round(ts[j] * (len(source_centers) - 1)), p)
    rows = []
    for j, (c, radius) in enumerate(zip(centers, radii)):
        tangent = unit(centers[min(j + 1, len(centers) - 1)] - centers[max(0, j - 1)])
        side = unit(np.cross([0, 1, 0], tangent))
        front = unit(np.cross(tangent, side))
        row = []
        for k in range(segments + 1):
            angle = TAU * k / segments
            t = ts[j]
            wrinkle = 1 + folds * math.sin(t * 37 + 3 * angle) * math.sin(math.pi * t)
            p = c + side * (radius[0] * math.cos(angle) * wrinkle) + front * (radius[1] * math.sin(angle) * wrinkle)
            uv = atlas(quadrant, k / segments, 1 - t) if quadrant is not None else (0, 0)
            w = weights(j, p) if callable(weights) else weights
            row.append((p, w))
        rows.append(row)
    for j in range(len(rows) - 1):
        for k in range(segments):
            middle = (centers[j] + centers[j + 1]) * .5
            out = rows[j][k][0] - middle
            if cutout and cutout(np.mean([rows[j][k][0], rows[j+1][k+1][0]], axis=0)):
                continue
            ids = []
            tile_u = math.floor((k + .5) / segments * repeat[0])
            tile_v = math.floor((ts[j] + ts[j + 1]) * .5 * repeat[1])
            for jj, kk in [(j, k), (j + 1, k), (j + 1, k + 1), (j, k + 1)]:
                p, w = rows[jj][kk]
                uv = atlas(quadrant, kk / segments * repeat[0] - tile_u, 1 - (ts[jj] * repeat[1] - tile_v)) if quadrant is not None else (0, 0)
                ids.append(surf.vertex(p, uv, color, w))
            surf.quad(*ids, outward=out)
    if caps:
        for j, direction in [(0, centers[0] - centers[1]), (-1, centers[-1] - centers[-2])]:
            w = weights(j % len(centers), centers[j]) if callable(weights) else weights
            idx = surf.vertex(centers[j], atlas(quadrant, .5, .5) if quadrant is not None else (0, 0), color, w)
            for k in range(segments):
                a = surf.vertex(rows[j][k][0], atlas(quadrant, .5 + .45 * math.cos(TAU*k/segments), .5 + .45 * math.sin(TAU*k/segments)) if quadrant is not None else (0, 0), color, w)
                b = surf.vertex(rows[j][k+1][0], atlas(quadrant, .5 + .45 * math.cos(TAU*(k+1)/segments), .5 + .45 * math.sin(TAU*(k+1)/segments)) if quadrant is not None else (0, 0), color, w)
                surf.tri(idx, a, b, direction)


def ellipsoid(surf, center, radius, weights, color=(1, 1, 1, 1), quadrant=None, latitudes=14, segments=24, uv_func=None):
    center = np.asarray(center)
    rows = []
    for j in range(latitudes + 1):
        t = math.pi * (.0001 + .9998 * j / latitudes)
        row = []
        for k in range(segments + 1):
            a = TAU * k / segments
            p = center + np.asarray(radius) * [math.sin(t) * math.sin(a), -math.sin(t) * math.cos(a), math.cos(t)]
            uv = uv_func(p) if uv_func else atlas(quadrant, k / segments, j / latitudes) if quadrant is not None else (0, 0)
            row.append(surf.vertex(p, uv, color, weights))
        rows.append(row)
    for j in range(latitudes):
        for k in range(segments):
            out = np.asarray(surf.p[rows[j][k]]) - center
            surf.quad(rows[j][k], rows[j + 1][k], rows[j + 1][k + 1], rows[j][k + 1], out)


def ribbon(surf, points, width, weights, quadrant=None, color=(1, 1, 1, 1), depth=.001):
    # Thin raised seams, button plackets, folded edges and laces in real geometry.
    pts = np.asarray(points)
    rows = []
    for i, p in enumerate(pts):
        tangent = unit(pts[min(i + 1, len(pts) - 1)] - pts[max(0, i - 1)])
        side = unit(np.cross(tangent, [0, -1, 0]))
        row = []
        for k in (-1, 1):
            pos = p + k * width * .5 * side + [0, -depth, 0]
            uv = atlas(quadrant, (k + 1) / 2, i / max(1, len(pts) - 1)) if quadrant is not None else (0, 0)
            row.append(surf.vertex(pos, uv, color, weights(pos) if callable(weights) else weights))
        rows.append(row)
    for i in range(len(rows) - 1):
        surf.quad(rows[i][0], rows[i][1], rows[i + 1][1], rows[i + 1][0], [0, -1, 0])


def make_head(face, cloth, skin, hair_surface):
    face_pixels = np.asarray(Image.open(SOURCE / "face-reference.png").convert("RGB")) / 255
    lm = np.asarray(json.loads((SOURCE / "face-landmarks.json").read_text()))[:468]
    scale = .43
    angle = math.atan2(-(lm[263, 1] - lm[33, 1]), lm[263, 0] - lm[33, 0])
    cs, sn = math.cos(angle), math.sin(angle)
    cx = (lm[33, 0] + lm[263, 0]) / 2
    cy = lm[168, 1]

    def plane(u, v):
        dx, up = u - cx, cy - v
        return dx * cs + up * sn, up * cs - dx * sn

    chin = plane(lm[152, 0], lm[152, 1])[1]

    def project(p):
        x, up = p[0] / scale, (p[2] - 1.502) / scale + chin
        return (cx + x * cs - up * sn, cy - (x * sn + up * cs))

    positions = []
    for u, v, z in lm:
        x, up = plane(u, v)
        p = (x * scale, (z - .228) * scale * .76, 1.502 + (up - chin) * scale)
        positions.append(p)
        face.vertex(p, (u, v), weights={5: 1})
    canonical_faces = []
    for line in (SOURCE / "canonical_face_model.obj").read_text().splitlines():
        if line.startswith("f "):
            f = [int(v.split("/")[0]) - 1 for v in line.split()[1:]]
            for i in range(1, len(f) - 1):
                canonical_faces.append([f[0], f[i], f[i + 1]])
    p = np.asarray(positions)
    f = np.asarray(canonical_faces)
    flip = np.cross(p[f[:, 1]] - p[f[:, 0]], p[f[:, 2]] - p[f[:, 0]]).sum(axis=0)[1] > 0
    for a, b, c in canonical_faces:
        face.tri(a, c, b) if flip else face.tri(a, b, c)

    # Canonical topology has openings at the eyelids and lips. Close these with
    # the same photographic UV coordinates so the eye sockets are not empty.
    loops = [
        [33, 7, 163, 144, 145, 153, 154, 155, 133, 173, 157, 158, 159, 160, 161, 246],
        [263, 249, 390, 373, 374, 380, 381, 382, 362, 398, 384, 385, 386, 387, 388, 466],
        [78, 95, 88, 178, 87, 14, 317, 402, 318, 324, 308, 415, 310, 311, 312, 13, 82, 81, 80, 191],
    ]
    for loop in loops:
        p = np.asarray([positions[i] for i in loop]).mean(axis=0)
        idx = face.vertex(p, project(p), weights={5: 1})
        for i, a in enumerate(loop):
            face.tri(idx, a, loop[(i + 1) % len(loop)], [0, -1, 0])

    # Extend the measured oval to the actual photographed hair silhouette.
    # One continuous loft owns the forehead and hair, avoiding intersecting caps.
    brow = [127,162,21,54,103,67,109,10,338,297,332,284,251,389,356]
    silhouette = [(.204,.43),(.190,.36),(.193,.278),(.22,.19),(.268,.117),
                  (.33,.063),(.398,.028),(.481,.02),(.57,.034),(.65,.069),
                  (.715,.129),(.753,.21),(.774,.30),(.775,.38),(.750,.44)]
    brow_rows = [brow]
    outer_positions = np.asarray(positions).copy()
    for j in range(1,13):
        t=j/12
        row=[]
        for idx,(u,v) in zip(brow,silhouette):
            original_uv=lm[idx,:2]
            uv=original_uv*(1-t)+np.array([u,v])*t
            x,up=plane(*uv)
            y=positions[idx][1]*(1-t)+.023*t-.052*math.sin(math.pi*t)
            pos=[x*scale,y,1.502+(up-chin)*scale]
            row.append(face.vertex(pos,uv.tolist(),weights={5:1}))
            if j==12:
                outer_positions[idx]=pos
        brow_rows.append(row)
    for j in range(12):
        for k in range(len(brow)-1):
            face.quad(brow_rows[j][k],brow_rows[j][k+1],brow_rows[j+1][k+1],brow_rows[j+1][k],[0,-1,0])
    face.tri(127, brow_rows[-1][0], 234, [0,-1,0])
    face.tri(356, 454, brow_rows[-1][-1], [0,-1,0])

    outline = [10,338,297,332,284,251,389,356,454,323,361,288,397,365,379,378,400,377,152,148,176,149,150,136,172,58,132,93,234,127,162,21,54,103,67,109]
    rings=[]
    for j in range(17):
        t=j/16*math.pi/2
        row=[]
        for idx in outline:
            x,y,z=outer_positions[idx]
            pos=[x*math.cos(t)+math.copysign(.005*math.sin(t*2),x),
                 y*(1-math.sin(t))+.112*math.sin(t),1.665+(z-1.665)*math.cos(t)]
            edge=project(outer_positions[idx])
            edge_color=face_pixels[int(np.clip(edge[1],0,.999)*len(face_pixels)),int(np.clip(edge[0],0,.999)*face_pixels.shape[1])]
            mix=min(1,j/6)
            rgb=edge_color*(1-mix)+np.array([.69,.46,.36])*mix
            row.append((pos,(*rgb,1)))
        rings.append(row)
    for j in range(16):
        for k in range(len(outline)):
            nxt=(k+1)%len(outline)
            mid=np.mean([rings[j][k][0],rings[j+1][nxt][0]],axis=0)
            is_hair=(outline[k] in brow and outline[nxt] in brow) or (mid[2]>1.614 and mid[1]>.045)
            target=hair_surface if is_hair else skin
            ids=[]
            for jj,kk in [(j,k),(j+1,k),(j+1,nxt),(j,nxt)]:
                pos,color=rings[jj][kk]
                uv=(kk/len(outline)*3,jj/16*1.3)
                ids.append(target.vertex(pos,uv,color=(1,1,1,1) if is_hair else color,weights={5:1}))
            target.quad(*ids,outward=mid-[0,0,1.665])

    # Ear shells, lobes and outer helix. The frontal surface has reference UVs.
    for sign in (-1, 1):
        ellipsoid(skin, [sign * .118, .001, 1.639], [.018, .017, .039], {5: 1}, color=SKIN, latitudes=18, segments=24)
        ellipsoid(skin, [sign * .119, -.015, 1.642], [.007, .004, .020], {5: 1}, color=(.59, .37, .30, 1), latitudes=12, segments=18)
        helix = []
        for j in range(28):
            a = -.9 + 5.1 * j / 27
            helix.append([sign * (.118 + .013 * math.cos(a)), -.018, 1.644 + .032 * math.sin(a)])
        tube(skin, helix, [[.0018, .0018]] * len(helix), {5: 1}, color=SKIN, segments=8)
    tube(skin, [[0, .015, 1.443], [0, .012, 1.487], [0, .01, 1.531], [0, .013, 1.595]],
         [[.068, .061], [.063, .057], [.059, .055], [.064, .06]],
         lambda j, p: blend(4, 5, (p[2] - 1.49) / .09), color=SKIN, segments=40)


def make_body(cloth, skin, details):
    # An elliptical, tailored shirt with a modest abdomen and shoulder slope.
    profile = [(.987, .175, .099), (1.009, .183, .108), (1.042, .196, .115),
               (1.086, .204, .121), (1.13, .208, .125), (1.178, .211, .125),
               (1.222, .213, .122), (1.268, .219, .117), (1.308, .225, .112),
               (1.347, .231, .107), (1.382, .228, .099), (1.412, .209, .090),
               (1.444, .173, .080), (1.47, .121, .068), (1.483, .067, .060)]
    tube(cloth, [[0, .008, z] for z, _, _ in profile], [[x, y] for _, x, y in profile],
         lambda j, p: torso_weights(p[2]), quadrant=0, segments=64, folds=.019, caps=False,
         cutout=lambda p: p[1] < -.025 and p[2] > 1.44 + abs(p[0]) * 1.1)

    def front_y(z, x=0):
        ry = float(np.interp(z, [p[0] for p in profile], [p[2] for p in profile]))
        rx = float(np.interp(z, [p[0] for p in profile], [p[1] for p in profile]))
        return .008 - ry * math.sqrt(max(.02, 1 - (x / rx) ** 2)) - .003

    placket = [[0, front_y(z), z] for z in np.linspace(1.012, 1.44, 26)]
    ribbon(cloth, placket, .018, lambda p: torso_weights(p[2]), quadrant=0, depth=.001)
    for z in np.linspace(1.04, 1.425, 7):
        ellipsoid(details, [0, front_y(z) - .004, z], [.004, .0018, .004], torso_weights(z), color=(.23, .20, .16, 1), latitudes=6, segments=10)
    # Folded collar leaves join the neck opening to two lowered points.
    for sign in (-1, 1):
        points = [[sign * .016, -.053, 1.484], [sign * .066, -.034, 1.487],
                  [sign * .108, -.079, 1.449], [sign * .041, -.107, 1.409]]
        ids = [cloth.vertex(p, atlas(0, u, v), weights={3: 1}) for p, (u, v) in zip(points, [(0, 0), (1, 0), (1, 1), (0, 1)])]
        cloth.quad(*ids, outward=[0, -1, 0])
        ribbon(details, [points[0], points[3], points[2]], .0017, {3: 1}, color=(.25, .28, .33, 1))
    # Small sewn breast pocket.
    pocket = [[.065, front_y(1.337, .065), 1.337], [.135, front_y(1.337, .135), 1.337],
              [.132, front_y(1.27, .132), 1.27], [.098, front_y(1.257, .098), 1.257],
              [.064, front_y(1.27, .064), 1.27]]
    center = np.mean(pocket, axis=0)
    def shirt_uv(p):
        u = (.75 + math.asin(p[0] / .225) / TAU) * 3 % 1
        v = 1 - ((p[2] - .987) / (1.483 - .987) * 2 % 1)
        return atlas(0, u, v)
    mid = cloth.vertex(center, shirt_uv(center), weights={3: 1})
    ids = [cloth.vertex(p, shirt_uv(p), weights={3: 1}) for p in pocket]
    for i in range(5):
        cloth.tri(mid, ids[i], ids[(i + 1) % 5], [0, -1, 0])
    ribbon(details, pocket + [pocket[0]], .0012, {3: 1}, color=(.47, .49, .52, 1), depth=.003)

    for sign, upper, lower, hand in [(-1, 6, 7, 8), (1, 9, 10, 11)]:
        shoulder = np.array([sign * .163, 0, 1.374])
        elbow = np.array([sign * .402, 0, 1.244])
        wrist = np.array([sign * .492, 0, 1.064])
        cuff = elbow * .68 + wrist * .32
        centers, radii = [], []
        for t in np.linspace(0, 1, 18):
            p = shoulder * (1 - t) + cuff * t
            r = float(np.interp(t, [0, .20, .5, 1], [.050, .082, .072, .054])) + .002 * math.sin(t * 6 * math.pi)
            centers.append(p)
            radii.append([r, r * .88])
        tube(cloth, centers, radii, lambda j, p: blend(3, upper, j / 3) if j < 3 else blend(upper, lower, (j / 17 - .62) / .30), quadrant=0, segments=32, folds=.04)
        tangent = unit(cuff - shoulder)
        tube(cloth, [cuff - tangent * .036, cuff - tangent * .03, cuff - tangent * .004, cuff],
             [[.061, .055], [.064, .057], [.064, .057], [.058, .052]], {lower: 1}, quadrant=0, segments=32)
        arm_centers = [cuff * (1 - t) + wrist * t for t in np.linspace(0, 1, 11)]
        arm_radii = [[.05 * (1 - t) + .028 * t + .007 * math.sin(math.pi * t), .044 * (1 - t) + .025 * t] for t in np.linspace(0, 1, 11)]
        tube(skin, arm_centers, arm_radii, lambda j, p: blend(lower, hand, max(0, (j / 10 - .85) / .3)), color=SKIN, segments=28)
        # Relaxed closed hand with individual fingers and a bent thumb.
        ellipsoid(skin, [sign * .500, -.003, 1.025], [.036, .024, .049], {hand: 1}, color=SKIN)
        for finger in range(4):
            x = sign * (.476 + finger * .016)
            z = 1.001 + abs(finger - 1.5) * .004
            pts = [[x, -.009, 1.033], [x + sign * .003, -.018, 1.006], [x + sign * .003, -.014, z - .017], [x, .003, z - .014]]
            tube(skin, pts, [[.009, .008], [.0095, .008], [.008, .007], [.006, .006]], {hand: 1}, color=SKIN, segments=10)
        tube(skin, [[sign * .474, .003, 1.046], [sign * .462, -.018, 1.025], [sign * .469, -.027, 1.008]],
             [[.013, .012], [.011, .010], [.009, .008]], {hand: 1}, color=SKIN, segments=12)

    # Denim pelvis overlaps the upper legs inside the cloth surface.
    tube(cloth, [[0, .012, z] for z in [.811, .845, .89, .947, .982, 1.004]],
         [[.132, .088], [.172, .099], [.187, .106], [.184, .106], [.174, .101], [.171, .099]],
         {1: 1}, quadrant=1, segments=48)
    for sign, upper, lower, foot in [(-1, 12, 13, 14), (1, 15, 16, 17)]:
        centers, radii = [], []
        for z in np.linspace(.935, .125, 29):
            x = sign * float(np.interp(z, [.13, .504, .936], [.138, .128, .104]))
            y = float(np.interp(z, [.13, .504, .936], [-.026, 0, .012]))
            r = float(np.interp(z, [.13, .20, .30, .40, .50, .58, .70, .83, .935], [.059, .061, .073, .073, .067, .077, .087, .091, .077]))
            centers.append([x, y, z])
            radii.append([r, r * .91])
        def leg_weights(j, p):
            if p[2] > .81:
                return blend(upper, 1, (p[2] - .81) / .21)
            return blend(upper, lower, (.575 - p[2]) / .14)
        tube(cloth, centers, radii, leg_weights, quadrant=1, segments=36, folds=.033)
        for side in (-1, 1):
            seam = [[c[0] + side * r[0] * .92, c[1] - r[1] * .39, c[2]] for c, r in zip(centers, radii)]
            ribbon(details, seam, .0011, lambda p: leg_weights(0, p), color=(.38, .31, .22, 1), depth=.0003)
        # Rounded leather shoes, raised vamp, separate sole and lace crossings.
        x = sign * .138
        sole_rows = [[x, -.038, z] for z in [.017, .022, .038, .044]]
        tube(details, sole_rows, [[.066, .135], [.068, .137], [.068, .137], [.065, .13]], {foot: 1}, color=(.085, .075, .065, 1), segments=40)
        shoe_rows = [[x, -.038 + (z - .05) * .35, z] for z in [.039, .05, .068, .09, .112, .131]]
        tube(cloth, shoe_rows, [[.066, .132], [.067, .133], [.064, .126], [.059, .107], [.05, .075], [.036, .045]], {foot: 1}, quadrant=2, segments=40)
        for z in [.086, .098, .108, .116]:
            y = -.065 + (z - .086) * 1.2
            ribbon(details, [[x - .027, y, z], [x + .026, y + .012, z + .007]], .003, {foot: 1}, color=(.12, .085, .06, 1), depth=.001)
    # Belt and buckle are separate from the shirt so their silhouette reads.
    tube(cloth, [[0, .012, z] for z in [.975, .978, 1.006, 1.01]],
         [[.178, .103], [.181, .106], [.179, .104], [.175, .1]], {1: 1}, quadrant=2, segments=64)
    buckle = [[-.026, -.098, .981], [.026, -.098, .981], [.026, -.098, 1.006], [-.026, -.098, 1.006], [-.026, -.098, .981]]
    ribbon(details, buckle, .004, {1: 1}, color=(.49, .42, .29, 1), depth=.007)
    ribbon(details, [[0, -.106, .98], [0, -.106, 1.007]], .003, {1: 1}, color=(.53, .47, .34, 1))
    # Curved pocket openings, fly stitching and raised belt loops.
    for sign in (-1, 1):
        ribbon(details, [[sign*x, y, z] for x,y,z in [(.075,-.085,.971),(.087,-.084,.944),(.109,-.079,.923),(.141,-.066,.915)]],
               .0016, {1:1}, color=(.38,.31,.23,1), depth=.006)
        ribbon(cloth, [[sign*.113,-.073,.961],[sign*.114,-.08,1.016]], .014, {1:1}, quadrant=1, depth=.008)
    ribbon(details, [[.008,-.098,.972],[.009,-.098,.905],[.015,-.092,.865]], .0012, {1:1}, color=(.35,.29,.23,1), depth=.004)


def read_glb(path):
    data = path.read_bytes()
    magic, version, size = struct.unpack_from("<4sII", data)
    assert (magic, version, size) == (b"glTF", 2, len(data))
    length, kind = struct.unpack_from("<II", data, 12)
    assert kind == 0x4E4F534A
    doc = json.loads(data[20:20 + length])
    offset = 20 + length
    binlength, kind = struct.unpack_from("<II", data, offset)
    assert kind == 0x004E4942
    return doc, bytearray(data[offset + 8:offset + 8 + binlength])


def build(output):
    original = ROOT / "assets/source/original-models/ivanov.glb"
    doc, binary = read_glb(original)
    original_doc = copy.deepcopy(doc)
    original_binary = bytes(binary)
    face = Surface("Ivanov identity / sculpted face and hairline", 0)
    cloth = Surface("Ivanov tailored clothes / hair / leather", 1)
    skin = Surface("Ivanov neck / forearms / hands", 3)
    details = Surface("Ivanov seams / buttons / shoes / buckle", 2)
    hair = Surface("Ivanov combed gray hair", 4)
    make_head(face, cloth, skin, hair)
    make_body(cloth, skin, details)

    def view(data, target=None):
        while len(binary) % 4:
            binary.append(0)
        offset = len(binary)
        binary.extend(data)
        item = {"buffer": 0, "byteOffset": offset, "byteLength": len(data)}
        if target:
            item["target"] = target
        doc["bufferViews"].append(item)
        return len(doc["bufferViews"]) - 1

    def accessor(values, dtype, gl_type, component, bounds=False, target=34962):
        arr = np.asarray(values, dtype=dtype)
        a = {"bufferView": view(arr.tobytes(), target), "componentType": component, "count": len(arr), "type": gl_type}
        if bounds:
            a["min"], a["max"] = arr.min(axis=0).tolist(), arr.max(axis=0).tolist()
        doc["accessors"].append(a)
        return len(doc["accessors"]) - 1

    primitives = []
    stats = []
    for surf in [face, cloth, skin, details, hair]:
        assert len(surf.p) < 65536
        assert np.isfinite(surf.p).all()
        assert np.allclose(np.asarray(surf.weights).sum(axis=1), 1)
        assert np.logical_and(np.asarray(surf.colors) >= 0, np.asarray(surf.colors) <= 1).all()
        attrs = {
            "POSITION": accessor(surf.p, "<f4", "VEC3", 5126, True),
            "NORMAL": accessor(surf.normals(), "<f4", "VEC3", 5126),
            "TEXCOORD_0": accessor(surf.uv, "<f4", "VEC2", 5126),
            "COLOR_0": accessor(surf.colors, "<f4", "VEC4", 5126),
            "JOINTS_0": accessor(surf.joints, "<u2", "VEC4", 5123),
            "WEIGHTS_0": accessor(surf.weights, "<f4", "VEC4", 5126),
        }
        indices = accessor(np.asarray(surf.faces).reshape(-1), "<u2", "SCALAR", 5123, target=34963)
        primitives.append({"attributes": attrs, "indices": indices, "material": surf.material, "mode": 4})
        stats.append({"surface": surf.name, "vertices": len(surf.p), "triangles": len(surf.faces)})
    doc["meshes"] = [{"name": "Ivanov reference reconstruction", "primitives": primitives}]
    doc["images"] = [{"name": name, "bufferView": view((SOURCE / name).read_bytes()), "mimeType": "image/png"} for name in ["face-reference.png", "materials.png", "hair.png"]]
    doc["samplers"] = [{"magFilter": 9729, "minFilter": 9987, "wrapS": 33071, "wrapT": 33071}, {"magFilter":9729, "minFilter":9987, "wrapS":10497, "wrapT":10497}]
    doc["textures"] = [{"sampler": 1 if i==2 else 0, "source": i} for i in range(3)]
    doc["materials"] = []
    for i, name in enumerate(["Measured face", "Cotton denim leather atlas", "Solid details", "Skin albedo", "Combed gray hair"]):
        pbr = {"baseColorFactor": [1, 1, 1, 1], "metallicFactor": 0, "roughnessFactor": .82}
        if i != 2:
            pbr["baseColorTexture"] = {"index": {0:0, 1:1, 3:0, 4:2}[i], "texCoord": 0}
        doc["materials"].append({"name": name, "pbrMetallicRoughness": pbr, "doubleSided": True})
    doc.setdefault("extras", {})["fightogm_art"] = {
        "version": 1, "coordinate_system": "Z-up, front -Y", "source": "Ivanov.jpg and ivanov_model_sheet.png",
        "builder": "scripts/art/build_ivanov.py", "rig_and_animations": "unchanged from supplied GLB",
    }
    while len(binary) % 4:
        binary.append(0)
    doc["buffers"] = [{"byteLength": len(binary)}]
    encoded = json.dumps(doc, ensure_ascii=False, separators=(",", ":")).encode("utf-8")
    encoded += b" " * (-len(encoded) % 4)
    out = struct.pack("<4sII", b"glTF", 2, 28 + len(encoded) + len(binary))
    out += struct.pack("<II", len(encoded), 0x4E4F534A) + encoded
    out += struct.pack("<II", len(binary), 0x004E4942) + binary
    for key in ("nodes", "skins", "animations"):
        assert original_doc[key] == doc[key], f"Changed {key}!"
    assert bytes(binary[:len(original_binary)]) == original_binary
    output.parent.mkdir(parents=True, exist_ok=True)
    output.write_bytes(out)
    report = {"output": str(output.relative_to(ROOT)), "original_sha256": hashlib.sha256(original.read_bytes()).hexdigest(),
              "output_sha256": hashlib.sha256(out).hexdigest(), "bytes": len(out), "surfaces": stats,
              "vertices": sum(s["vertices"] for s in stats), "triangles": sum(s["triangles"] for s in stats),
              "embedded_textures": len(doc["images"]), "preserved_clips": [a["name"] for a in doc["animations"]],
              "rig_and_animation_data_unchanged": True}
    (SOURCE / "build-report.json").write_text(json.dumps(report, ensure_ascii=False, indent=2), encoding="utf-8")
    print(json.dumps(report, ensure_ascii=True, indent=2))


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output", type=Path, default=ROOT / "assets/models/ivanov.glb")
    build(parser.parse_args().output.resolve())
