# Исходники графики FIGHTOGM 2D

## Новацкий с чёрным шнеком — 25 сентября 2026

Переработаны все три листа `assets/sprites/novatskiy/`: движение, атаки и реакции,
36 поз. Во всех позах Новацкий держит чёрный промышленный шнек по
`assets/references/novatskiy/Shnek.jpg`; лицо и одежда сохранены по прежним спрайтам.
Использован встроенный image_gen. Исходный референс оставлен без изменений.
Точные запросы, исходные изображения и принятые результаты записаны в
[novatskiy-auger-prompts.json](novatskiy-auger-prompts.json).
Лист атак дополнительно исправлен генератором: фигуры разнесены, чтобы оружие
не касалось соседнего бойца в атласе, уточнена поза возврата после высокого удара ногой.
PNG не обрабатывались скриптами; метаданные рамок и масштаба созданы `build_roster.py`.
Особый приём использует три фазы ближнего удара, суперприём — три контакта шнеком.
Проверка оружия и отсутствия притяжки: `test-output/auger-dist/`.

## Цветков и Козерод — 25 сентября 2026

Добавлены два персонажа по новым фотографиям:
`assets/references/tsvetkov/Цветков.jpg` и `assets/references/kozerod/Козерод.jpg`.
Спрайты созданы встроенным image_gen, по три листа и 36 поз для каждого.
Готовые PNG находятся в `assets/sprites/tsvetkov/` и `assets/sprites/kozerod/`.
Точные принятые запросы, использованные референсы и пути исходников генератора:
[Цветков](tsvetkov-prompts.json), [Козерод](kozerod-prompts.json).

Генератор получил открытые в диалоге предпросмотры: прямое чтение путей было
недоступно из-за ошибки Windows sandbox. Исходные фотографии сохранены без изменений.
Референсы движения задавали лицо, одежду и стиль последующим атакам и реакциям.
Цветков сохраняет кожаную мотокуртку со светлыми полосами и джинсы;
Козерод — короткую светлую стрижку и чёрный костюм с круглой нашивкой.
Добавлены 72 позы и 34 клипа. Теперь в игре 12 бойцов, 432 позы и 205 клипов.
PNG не обрабатывались скриптами; границы, области отрисовки и масштабы записаны в JSON.
Проверка новых бойцов и всей сборки: `test-output/2d-expansion-dist/`.

Режим: встроенный image_gen (генерация и редактирование изображений OpenAI), 15 сентября 2026.

## Полный состав: девять новых бойцов

Добавлены Гаглоев, Киселик, Новацкий, Жирнов, Елхимов, Калачев, Федосеев, Халиман и Шуев.
У каждого по три новых PNG-листа: `assets/sprites/<id>/movement.png`, `attacks.png`,
`actions.png`; каждый лист содержит 12 поз. Итого добавлены 27 листов и 324 позы,
вместе с Ивановым — 360 поз и 171 анимация. Все листы созданы встроенным image_gen.

Полные точные запросы, пути референсов и оригиналов генерации всех 27 принятых
листов сохранены в [roster-all-prompts.json](roster-all-prompts.json).
Фотография каждого бойца задавала внешность, его model sheet — одежду;
готовые стойки Иванова задавали общий стиль, затем собственный лист movement
использовался для сохранения внешности в атаках и реакциях. Референсы лежат
в `assets/references/<id>/`; исходники генератора оставлены на исходных местах.

PNG используются без внешней обработки. `scripts/sprites/build_roster.py` читает
границы персонажей и записывает только JSON с кадрами, опорными точками и клипами.
При пересечении прямоугольников соседних поз JSON задаёт несколько областей
отрисовки, чтобы не показывать чужую ступню или голову в углу кадра.
Зелёный фон удаляет игровой шейдер; цветные снаряды и следы рисует native raylib.
Результаты проверки всех бойцов в готовой сборке: `test-output/2d-roster-dist/`.

## Иванов и арена

Референсы: `assets/references/ivanov/Ivanov.jpg` и `ivanov_model_sheet.png`.
Последующие позы использовали утверждённый лист `assets/sprites/ivanov/stances.png`.
Листы хранятся в `assets/sprites/ivanov/`, фон — `assets/backgrounds/ogm-workshop.png`.
Всего шесть листов, 36 поз, 18 игровых анимаций. Визуальные файлы созданы генератором.

Первая версия стойки содержала нарисованную шахматную подложку вместо прозрачности.
Она сохранена как `stances-checkerboard-rejected.png` и не используется в игре.
Фон исправлен тем же генератором на чистый зелёный; игровой шейдер удаляет его и
подавляет зелёные края. PNG не перекрашивались внешними скриптами.
`build_manifest.py` только читает пиксели и сохраняет прямоугольники/опоры в JSON.
Проверка реального OpenGL-рендера сохраняется в `test-output/2d-dist/`.

## Точные запросы

### stances.png

```text
Create a production-ready 2D arcade fighting-game SPRITE SHEET of Ivanov. Image 1 is the authoritative reference for his facial identity; image 2 is the wardrobe reference. Preserve this mature man's swept short gray hair, thick dark moustache, pronounced nose, age, face, and average sturdy build. Blue-gray small-plaid shirt with rolled sleeves, dark blue jeans, brown belt and brown leather shoes. Painterly realistic 2D videogame art with crisp readable edges and richly shaded fabric, restrained outlines, dramatic yet clear warm key lighting, consistent lighting throughout. This is drawn sprite art, not a 3D model render, not pixel art and not a photography contact sheet. All poses face SCREEN RIGHT, a side-view fighting stance with chest and face turned slightly toward the viewer so the identity reads. Anatomically natural martial-arts stance, knees flexed, fists guarding his face. SAME character, wardrobe, proportions, camera and scale in all frames.
STRICT LAYOUT: wide 3:2 canvas, exactly THREE COLUMNS and TWO ROWS, six equal rectangular cells in reading order. Each cell contains ONE entire body, no clipping or crossing cell edges. Generous transparent spacing around every character, at least 8% cell margin. Ground baseline is 90% down each cell. Consistent scale: standing man is 78% of cell height, crouching and jumping poses naturally shorter. Real transparent RGBA background everywhere around the sprites. NO background, floor, ground shadows, boxes, grid lines, text, labels, logos, weapons, effects or additional people. Do not draw a checkerboard. Output a clean transparent PNG sprite atlas at highest available resolution.
FRAMES in reading order: top left = neutral fighting guard, left foot forward, two fists raised; top middle = same guard breathing in, shoulders slightly raised, knees subtly flexed; top right = same guard breathing out with slight weight shift; bottom left = firm high block, both forearms protect face, grounded feet; bottom middle = deep low crouching guard, bent knees and one forearm protects head; bottom right = airborne jump pose with both knees bent and tucked slightly, guard up, same screen-right facing. No duplicates; retain identity exactly.
```

### stances.png — исправление фона

```text
Edit this exact sprite sheet. Preserve all six character poses, their facial identity, clothes, locations, sizes, details, and the 3-column by 2-row arrangement. Replace ONLY the entire gray-and-white checkerboard background with a perfectly uniform, flat, saturated pure chroma-key GREEN background, RGB 0,255,0 (#00FF00). Every pixel outside the six characters must be the same solid bright green, including gaps between fingers, arms, and legs. No gradients, no shadows, no checkerboard. Maintain clean sharp character boundaries. Do not add any text, grids, objects, effects or poses. The green will be keyed out by the native game's sprite shader.
```

### punches.png

```text
Create a production-ready 2D arcade fighting-game SPRITE SHEET of Ivanov. Image 1 is the authoritative reference for his facial identity; image 2 is the wardrobe reference. Preserve this mature man's swept short gray hair, thick dark moustache, pronounced nose, age, face, and average sturdy build. Blue-gray small-plaid shirt with rolled sleeves, dark blue jeans, brown belt and brown leather shoes. Painterly realistic 2D videogame art with crisp readable edges and richly shaded fabric, restrained outlines, dramatic yet clear warm key lighting, consistent lighting throughout. This is drawn sprite art, not a 3D model render, not pixel art and not a photography contact sheet. All poses face SCREEN RIGHT, a side-view fighting stance with chest and face turned slightly toward the viewer so the identity reads. Anatomically natural martial-arts stance, knees flexed, fists guarding his face. SAME character, wardrobe, proportions, camera and scale in all frames.
STRICT LAYOUT: wide 3:2 canvas, exactly THREE COLUMNS and TWO ROWS, six equal rectangular cells in reading order. Each cell contains ONE entire body, no clipping or crossing cell edges. Generous green spacing around every character, at least 8% cell margin. Ground baseline is 90% down each cell. Consistent scale: standing man is 78% of cell height, crouching and jumping poses naturally shorter. Flat uniform pure chroma-key GREEN (#00FF00) background around all characters. No other background colours. NO background, floor, ground shadows, boxes, grid lines, text, labels, logos, weapons, effects or additional people. Do not draw a checkerboard. Output a clean PNG sprite atlas at highest available resolution.

Image 3 is the APPROVED 2D sprite art: match its face, costume, drawing style and proportions exactly. Create six new ATTACK frames. Top row is one left jab: 1 windup with left fist at face and knees loaded, 2 left fist fully extended horizontally to SCREEN RIGHT at head height, straight arm, shoulder forward, 3 left arm retracting into guard. Bottom row is one heavy rear-hand cross: 4 torso twisting back to wind up the right fist, 5 forceful fully extended right punch to SCREEN RIGHT at face height with hips rotated, 6 right hand returning and weight settling into guard. Both feet visible and grounded in every frame. Clearly different windup, contact and recovery silhouettes. The extended fist stays entirely inside its cell. No motion blur or trails.
```

### walk.png

```text
Create a production-ready 2D arcade fighting-game SPRITE SHEET of Ivanov. Image 1 is the authoritative reference for his facial identity; image 2 is the wardrobe reference. Preserve this mature man's swept short gray hair, thick dark moustache, pronounced nose, age, face, and average sturdy build. Blue-gray small-plaid shirt with rolled sleeves, dark blue jeans, brown belt and brown leather shoes. Painterly realistic 2D videogame art with crisp readable edges and richly shaded fabric, restrained outlines, dramatic yet clear warm key lighting, consistent lighting throughout. This is drawn sprite art, not a 3D model render, not pixel art and not a photography contact sheet. All poses face SCREEN RIGHT, a side-view fighting stance with chest and face turned slightly toward the viewer so the identity reads. Anatomically natural martial-arts stance, knees flexed, fists guarding his face. SAME character, wardrobe, proportions, camera and scale in all frames.
STRICT LAYOUT: wide 3:2 canvas, exactly THREE COLUMNS and TWO ROWS, six equal rectangular cells in reading order. Each cell contains ONE entire body, no clipping or crossing cell edges. Generous green spacing around every character, at least 8% cell margin. Ground baseline is 90% down each cell. Consistent scale: standing man is 78% of cell height, crouching and jumping poses naturally shorter. Flat uniform pure chroma-key GREEN (#00FF00) background around all characters. No other background colours. NO background, floor, ground shadows, boxes, grid lines, text, labels, logos, weapons, effects or additional people. Do not draw a checkerboard. Output a clean PNG sprite atlas at highest available resolution.

Image 3 is the approved sprite art style. Create SIX chronological keyframes of a grounded combat shuffle walking toward SCREEN RIGHT, fists continuously raised and eyes looking right. Frame 1: front rightward foot steps forward while rear foot stays back; Frame 2: weight passes onto forward foot and rear heel lifts; Frame 3: rear foot slides closer under hips, knees bent; Frame 4: rear foot starts another short step while lead foot anchors; Frame 5: weight settles and rear foot extends back; Frame 6: return to ready combat stance matching frame 1. Natural alternating legs and visible foot motion, not six identical poses. Keep torso upright, same size and same head height, no motion blur, no effects. Entire feet must fit each cell. Chroma green background (#00FF00), never a checkerboard.
```

### kicks.png

```text
Create a production-ready 2D arcade fighting-game SPRITE SHEET of Ivanov. Image 1 is the authoritative reference for his facial identity; image 2 is the wardrobe reference. Preserve this mature man's swept short gray hair, thick dark moustache, pronounced nose, age, face, and average sturdy build. Blue-gray small-plaid shirt with rolled sleeves, dark blue jeans, brown belt and brown leather shoes. Painterly realistic 2D videogame art with crisp readable edges and richly shaded fabric, restrained outlines, dramatic yet clear warm key lighting, consistent lighting throughout. This is drawn sprite art, not a 3D model render, not pixel art and not a photography contact sheet. All poses face SCREEN RIGHT, a side-view fighting stance with chest and face turned slightly toward the viewer so the identity reads. Anatomically natural martial-arts stance, knees flexed, fists guarding his face. SAME character, wardrobe, proportions, camera and scale in all frames.
STRICT LAYOUT: wide 3:2 canvas, exactly THREE COLUMNS and TWO ROWS, six equal rectangular cells in reading order. Each cell contains ONE entire body, no clipping or crossing cell edges. Generous green spacing around every character, at least 8% cell margin. Ground baseline is 90% down each cell. Consistent scale: standing man is 78% of cell height, crouching and jumping poses naturally shorter. Flat uniform pure chroma-key GREEN (#00FF00) background around all characters. No other background colours. NO background, floor, ground shadows, boxes, grid lines, text, labels, logos, weapons, effects or additional people. Do not draw a checkerboard. Output a clean PNG sprite atlas at highest available resolution.

Image 3 is the approved sprite art style. Create six ATTACK animation keyframes. TOP ROW: 1 high kick windup, knee raised in front, both hands guard head, back foot grounded; 2 fully extended horizontal side kick toward SCREEN RIGHT at opponent's chest/head height, brown boot farthest right, torso leaning back, one supporting leg bent, full body fits cell; 3 leg retracting into a chamber after the kick, fists guarding. BOTTOM ROW: 4 low sweeping kick windup, man crouching and lowering weight on rear leg; 5 low sweeping kick extended toward SCREEN RIGHT near ankle height, low crouch with supporting bent leg, arms counterbalancing; 6 low kick recovery, front leg returning under body into crouched guard. Natural human anatomy, exactly two arms and two legs. Each sprite individually fully inside its equal cell with ample green margin. No blur, trails, effects or floor shadows. Uniform pure bright green #00FF00 background.
```

### reactions.png

```text
Create a production-ready 2D arcade fighting-game SPRITE SHEET of Ivanov. Image 1 is the authoritative reference for his facial identity; image 2 is the wardrobe reference. Preserve this mature man's swept short gray hair, thick dark moustache, pronounced nose, age, face, and average sturdy build. Blue-gray small-plaid shirt with rolled sleeves, dark blue jeans, brown belt and brown leather shoes. Painterly realistic 2D videogame art with crisp readable edges and richly shaded fabric, restrained outlines, dramatic yet clear warm key lighting, consistent lighting throughout. This is drawn sprite art, not a 3D model render, not pixel art and not a photography contact sheet. All poses face SCREEN RIGHT, a side-view fighting stance with chest and face turned slightly toward the viewer so the identity reads. Anatomically natural martial-arts stance, knees flexed, fists guarding his face. SAME character, wardrobe, proportions, camera and scale in all frames.
STRICT LAYOUT: wide 3:2 canvas, exactly THREE COLUMNS and TWO ROWS, six equal rectangular cells in reading order. Each cell contains ONE entire body, no clipping or crossing cell edges. Generous green spacing around every character, at least 8% cell margin. Ground baseline is 90% down each cell. Consistent scale: standing man is 78% of cell height, crouching and jumping poses naturally shorter. Flat uniform pure chroma-key GREEN (#00FF00) background around all characters. No other background colours. NO background, floor, ground shadows, boxes, grid lines, text, labels, logos, weapons, effects or additional people. Do not draw a checkerboard. Output a clean PNG sprite atlas at highest available resolution.

Use the attached standing sprites as the exact art, face and outfit continuity reference. Six REACTION poses in reading order: 1) recoiling backwards from a chest hit, face recoils left, still on feet; 2) falling backwards, knees buckling, torso tilted left; 3) lying knocked down horizontally on his back, head LEFT and feet RIGHT, whole body fits comfortably inside one cell; 4) getting up on one knee, hands pushing up; 5) nearly standing up with guard returning; 6) victory pose standing tall, one clenched fist raised high. No blood, wounds or effects. Full figure in every cell. Keep consistent anatomical dimensions even when the bounding box height changes for lying poses.
```

### specials.png

```text
Create a production-ready 2D arcade fighting-game SPRITE SHEET of Ivanov. Image 1 is the authoritative reference for his facial identity; image 2 is the wardrobe reference. Preserve this mature man's swept short gray hair, thick dark moustache, pronounced nose, age, face, and average sturdy build. Blue-gray small-plaid shirt with rolled sleeves, dark blue jeans, brown belt and brown leather shoes. Painterly realistic 2D videogame art with crisp readable edges and richly shaded fabric, restrained outlines, dramatic yet clear warm key lighting, consistent lighting throughout. This is drawn sprite art, not a 3D model render, not pixel art and not a photography contact sheet. All poses face SCREEN RIGHT, a side-view fighting stance with chest and face turned slightly toward the viewer so the identity reads. Anatomically natural martial-arts stance, knees flexed, fists guarding his face. SAME character, wardrobe, proportions, camera and scale in all frames.
STRICT LAYOUT: wide 3:2 canvas, exactly THREE COLUMNS and TWO ROWS, six equal rectangular cells in reading order. Each cell contains ONE entire body, no clipping or crossing cell edges. Generous green spacing around every character, at least 8% cell margin. Ground baseline is 90% down each cell. Consistent scale: standing man is 78% of cell height, crouching and jumping poses naturally shorter. Flat uniform pure chroma-key GREEN (#00FF00) background around all characters. No other background colours. NO background, floor, ground shadows, boxes, grid lines, text, labels, logos, weapons, effects or additional people. Do not draw a checkerboard. Output a clean PNG sprite atlas at highest available resolution.

Use the attached sprite sheet as the exact continuity reference. Six SPECIAL ACTION poses in reading order: 1) standing with one hand cupped to his ear, comically straining to hear, the other fist guarding; 2) second frame of that listening pose with a slight head lean; 3) grappling preparation leaning forward, both hands open and reaching SCREEN RIGHT at chest height; 4) grappling contact, both arms fully extended SCREEN RIGHT ready to grab an opponent, no opponent drawn; 5) power charging stance, fists clenched at his hips, determined face; 6) same power stance tensing up with fists a little farther from his body. No power effects, halos, glows, shadows or props. Full body in every cell with feet on a consistent baseline, clear green space between sprites.
```

### ogm-workshop.png

```text
Create a finished 16:9 widescreen 2D fighting game arena BACKGROUND plate, 1536x1024 or preferably wide 1536x864 if possible. A Soviet-era heavy industrial maintenance workshop, department of the chief mechanic, empty after the shift. Strict eye-level SIDE VIEW fighting-game composition: broad continuous horizontal concrete and steel floor across the entire width at the bottom quarter, the fighter ground line will be 83 percent down the screen. Camera looks at the long back wall straight on, wall verticals straight, modest perspective only in background machinery. No diagonal perspective fighting lane, no central vanishing corridor. Huge frosted gridded factory windows, old steel columns, overhead crane beam, pipes, workbenches, an old milling machine at far left and red tool cabinets at far right. Clear uncluttered central fighting area. Painterly detailed realistic 2D arcade art, sophisticated dark charcoal and desaturated teal palette with pools of warm amber tungsten light and cool blue night window light. Rich believable materials and cinematic atmosphere, a little haze, subtle grunge, beautiful contrast but no extreme darkness. Lighting matches hand-painted realistic fighter sprites in gray-blue plaid shirts. Background machinery is subdued so two large fighters remain readable. No people, characters, creatures, text, lettering, logos, UI, health bars, borders, watermarks, green screen or pixel art. Full bleed environment image, production-ready 2D game backdrop.
```
