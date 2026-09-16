# FightOGM — special/ultimate animated GLBs

This pack replaces the previous `FightOGM_animated_models` files.

## Install

Copy all 10 `.glb` files into:

`FightOGM/assets/models/`

and overwrite the previous versions.

Recommended final structure:

FightOGM/
  assets/
    models/
      ivanov.glb
      gagloev.glb
      kiselik.glb
      novatskiy.glb
      zhirnov.glb
      elkhimov.glb
      kalachev.glb
      fedoseev.glb
      khaliman.glb
      shuev.glb
    references/
      ...original photos and model sheets...
    audio/
    arenas/
    ui/

Also copy:
`special_animation_manifest.json`
to:
`FightOGM/assets/models/special_animation_manifest.json`

## Common embedded animations

idle
walk
crouch
jump
punch_light
punch_heavy
kick_high
block
hit_react
knockdown
get_up
grab
victory
special_pose

## Character-specific embedded animations

Ivanov:
bad_hearing
production_invisible
omni_transform
omni_rush

Gagloev:
skewer_control
shashlik

Kiselik:
sueta_dash
yes_boss
everything_at_once

Novatskiy:
screw_throw
snap_back
screw_batch

Zhirnov:
ai_agent
autopilot
do_it_for_me

Elkhimov:
wrench_32
repair_armor
overhaul

Kalachev:
water_burst
was_not_here
break_return

Fedoseev:
sarcasm_wave
slow_clap
genius_plan

Khaliman:
plc_pulse
standby
gym_mode
asutp

Shuev:
karate_flurry
stance_counter
endless_shift

## Important implementation note

The animation clips only animate the fighter skeleton.
Props and VFX mentioned in the manifest (skewer, wrench, auger, AI agent,
water wave, PLC panels, Omni aura, smoke, etc.) should be spawned and controlled by Go code.

Do not try to extract gameplay hit timing from the visuals automatically.
Define hitboxes, active frames, damage and VFX trigger times in Go.

Recommended code-side event mapping:

animation start -> lock fighter state
windup marker -> spawn prop/VFX if needed
active window -> enable hitbox
impact -> hit-stop / camera shake / damage
animation end -> return to idle

## raylib-go note

Load with `LoadModel()` / `LoadModelAnimations()` and select animations by name if your wrapper exposes names.
If animation names are not exposed conveniently by the raylib binding,
load `special_animation_manifest.json` and keep a deterministic index table in Go.

The safest production approach is:
1. enumerate animations once at startup;
2. cache name -> animation index;
3. never hardcode an index without verifying it.

## Prototype quality

These remain prototype skinned models with rigid body-part weights.
They are suitable for gameplay integration and animation state-machine work,
but final production deformation should use smooth humanoid weights.
