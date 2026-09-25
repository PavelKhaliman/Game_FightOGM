# Сторонние компоненты

- [raylib-go v0.55.1](https://github.com/gen2brain/raylib-go), авторы проекта,
  лицензия zlib. Включает raylib и его зависимости с их исходными уведомлениями.
- [raylib](https://www.raylib.com/), Copyright (c) 2013-2024 Ramon Santamaria,
  лицензия zlib / libpng.
- [Go fonts](https://go.googlesource.com/image/+/refs/tags/v0.24.0/font/gofont/),
  Copyright (c) 2016 by Bigelow & Holmes Inc. Все права защищены.
  Шрифты Go Regular и Go Bold встроены без изменений.
- golang.org/x/image, golang.org/x/sys, golang.org/x/exp — The Go Authors,
  лицензия BSD-3-Clause; github.com/ebitengine/purego — Apache-2.0.
- Исходные GLB и референсы предоставлены владельцем проекта; их авторство
  и права на использование не изменяются этой реализацией.
- [MediaPipe](https://github.com/google-ai-edge/mediapipe), The MediaPipe Authors,
  Apache-2.0. В архивном 3D-инструменте топология `canonical_face_model.obj` использована для геометрии
  лица Иванова; координаты, UV, закрытие глаз/рта и продолжение головы изменены.
  Face Landmarker применялся только при подготовке модели. Python и MediaPipe
  не входят в исполняемый файл игры. Лицензия: `licenses/mediapipe-LICENSE.txt`.

Полные тексты лицензий компонентов находятся в `licenses/` и в исходниках
зафиксированных Go-модулей. Фотографии и референсы в распространяемую сборку
не копируются. Текущая 2D-игра не загружает модели GLB.

2D-спрайты всех двенадцати бойцов и фон цеха созданы встроенным генератором изображений OpenAI.
Спрайты подготовлены по предоставленным референсам. Запросы и происхождение
сохранены в assets/source/2d/README.md. Python/NumPy/Pillow используются только
для анализа границ кадров при подготовке JSON, в игру не входят.
