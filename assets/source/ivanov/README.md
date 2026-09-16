# Иванов: модель по референсам

## Что включено

- `../../models/ivanov.glb` — готовая объёмная модель с пятью частями и тремя встроенными PNG.
- `face-reference.png` — фронтальная текстура, подготовленная по `Ivanov.jpg` и модельному листу.
- `materials.png` — атлас клетчатой ткани, джинсовой ткани, кожи и волос. Для затылка используется отдельная текстура ниже.
- `hair.png` — материал седых волос для боковых и задних поверхностей головы.
- `face-landmarks.json` — 478 измеренных точек; первые 468 задают геометрию лица.
- `canonical_face_model.obj` — исходная топология MediaPipe, Apache-2.0.
- `build-report.json` — количество вершин/треугольников, SHA-256 и перечень сохранённых клипов.
- `../original-models/ivanov.glb` — неизменённая резервная копия.

Голова имеет объёмный нос, подбородок, лоб, затылок и уши. Волосы продолжают
контур головы; рубашка, воротник, манжеты, руки, пальцы, джинсы, пояс и обувь
собраны на исходном скелете. Это реконструкция по изображениям, а не 3D-скан:
сходство спереди выше, чем точность затылка и профиля. Мимика остаётся нейтральной.

Из исходного GLB без изменений сохраняются `nodes`, `skins`, `animations`
и исходная BIN-секция. Новая геометрия и текстуры добавляются в конец данных,
а ссылка на отображаемые поверхности заменяется. Все 18 клипов работают с новым телом.
Игрокам Python, MediaPipe и исходные фотографии не нужны.

## Повторная сборка

Python 3.13, NumPy 2.2.6 и Pillow 11.2.1. Из корня проекта:

```powershell
python -m pip install -r scripts/art/requirements.txt
python scripts/art/build_ivanov.py
go run ./cmd/modelreview -compare assets/source/original-models/ivanov.glb
.\build.bat
```

В текущем рабочем окружении Python с зависимостями находится в
`.tools/art/venv/Scripts/python.exe`. Повторная сборка использует сохранённые
измерения, не обращается к сети и не запускает генерацию изображений.

Для повторного измерения лица есть `scripts/art/extract_face.py`.
Требуется отдельная установка MediaPipe 0.10.35 с зависимостями и
[официальная модель Face Landmarker](https://storage.googleapis.com/mediapipe-models/face_landmarker/face_landmarker/float16/1/face_landmarker.task).
[Документация MediaPipe](https://developers.google.com/edge/mediapipe/solutions/vision/face_landmarker/python).
Топология взята из
[canonical_face_model.obj](https://github.com/google-ai-edge/mediapipe/blob/master/mediapipe/modules/face_geometry/data/canonical_face_model.obj).

## Проверка

`cmd/modelreview` сохраняет реальные кадры из raylib/OpenGL в `test-output/ivanov`:
общий вид спереди, 3/4, сбоку, сзади, три крупных ракурса лица и `comparison.png`.
Это рендеры GLB, а не сгенерированные изображения результата.

`go run ./cmd/fightogm -smoke-test -output test-output/ivanov-game`
проверяет все части всех моделей и 171 исходную анимацию, игровые переходы,
зеркальный бой и суперприёмы.

## Генерация текстур

Применён навык **imagegen**, режим **встроенный image_gen**, без CLI/API fallback.
Все выбранные PNG скопированы в эту папку и встроены в игровой GLB.
Генерация создала только растровые материалы; геометрия и привязка к скелету
собраны отдельно скриптом проекта.

### face-reference.png

Референсы: `assets/references/ivanov/Ivanov.jpg` (личность) и
`assets/references/ivanov/ivanov_model_sheet.png` (дополнительный образ).

```text
Create a photorealistic head texture reference photograph for a 3D videogame character. Image 1 is the authoritative identity reference of the person; Image 2 is the supplementary character sheet. Preserve the exact identity from Image 1: mature man, wide prominent straight nose, dense dark moustache, gray hair combed to his right, pronounced brow, narrow eyes, broad squarish chin, natural asymmetry and age. Produce ONE perfectly straight-on orthographic passport-view head and neck on a flat warm gray background, entirely visible hair, both ears and chin, bottom edge at collarbone with a little gray-blue plaid shirt collar. Head absolutely upright, gaze straight toward viewer, mouth closed neutral expression, no smile. Face centered horizontally at 50% of width; head including hair occupies 85% of image height, ears span 68% of image width. Square canvas. Absolutely flat diffuse shadow-free texture-capture lighting, no side lighting, no cast shadows, minimal specular highlights, no beauty retouching. Detailed real skin, eyebrows, eyelids, individual moustache hairs and gray hair. The person's real facial identity and age are the highest priority, not beautification. No text, labels, frames, extra portraits, props, graphic layout or stylization. This image will be mapped onto real 3D geometry, so facial details must be crisp, symmetric perspective, and free of hard baked shadows.
```

### materials.png

Те же два референса, одежда и тип волос.

```text
Generate a square game material texture atlas with exactly FOUR equal square quadrants, edge-to-edge with no gaps, no labels, no borders. Materials are derived from the clothes and gray hair of the reference character. TOP LEFT: very fine blue-gray and muted lavender plaid woven cotton shirt fabric matching the reference man's shirt, flat unfolded fabric, repeating small check pattern at uniform scale, no buttons or collars or objects. TOP RIGHT: worn dark indigo blue denim fabric matching his jeans, fine realistic diagonal weave, subtle wear, flat seamless-looking material. BOTTOM LEFT: dark warm brown leather for his belt and brown shoes, subtle natural leather grain, uniform diffuse albedo, no large objects or stitching. BOTTOM RIGHT: closely combed short salt-and-pepper gray hair texture, predominantly silvery gray with some dark strands, realistic thin parallel gently curving hair strands, no scalp face or other objects. Orthographic material scans, flat diffuse color only, consistent soft lighting, NO cast shadows, NO 3D perspective, NO clothing silhouettes, NO rendered garments, NO text. This must be a usable albedo material atlas, not a presentation board. High detail 2048x2048.
```

### hair.png

Референс цвета и типа волос: `face-reference.png`.

```text
Create a production-ready seamless albedo material texture for the short gray hair of this mature man. The supplied photograph is a COLOR AND HAIR TYPE reference only. Fill the ENTIRE square canvas edge-to-edge with very fine closely combed salt-and-pepper hair, realistic individual silvery-gray hairs with darker charcoal roots, muted slightly warm gray. Uniform fine scale throughout, hundreds of thin small parallel strands flowing gently diagonally from upper left to lower right. Like a flat scanned close-cropped hair material intended to wrap the sides and back of a realistic 3D head. Soft diffuse illumination, no highlights baked into big shapes, no shadow gradients. Absolutely NO face, scalp skin, forehead, eyes, ears, hairline, head silhouette, swirls, loops, braids, fur clumps, text, labels, borders or atlas divisions. Avoid coarse zebra-like white and black waves: fine natural low-contrast hair detail only. Square texture, 1024x1024.
```

