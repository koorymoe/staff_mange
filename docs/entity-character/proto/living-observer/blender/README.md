# Local 3D character study — not approved production art

This original procedural character is a **new approximation**, not a recovered
3D version of the approved raster face. No employee image was sent anywhere.
Built locally with Blender 5.2.1; no new project dependency or paid service.

Open `output/amani-3d-study.blend`. The 180-frame timeline is 7.5 seconds:
front idle, whole-body right turn, rightward walk, stop, return to front.
Press Space over the viewport to play. The camera is fixed for judging direction.
`output/turn-and-walk.gif` is a lighter studio-shaded motion review;
`front.png` and `side-walk.png` are separately lit renders of the same mesh.

## What is implemented

- Real closed 3D surfaces for the head, clothing and shoes; volumetric eyes,
  nose, ears, beard and hair. Blue vest and black clothes.
- Object-joint hierarchy with keyed arm movement and two-link leg IK baked to
  mesh transforms. **Not a skinned armature** or production facial rig yet.
- Direction is changed before travel. Mid-cycle stance feet are fixed in
  world space. Start/stop blends are approximate and still need visual review.
- Editable named objects, materials, camera, lights and keyframes in the Blend.
- The small circular badge is a placeholder, not the official company logo.

## Not implemented / not claimed

Reference-face fidelity, detailed hands, blinking, talking, paper handoff,
backward/leftward journeys, production GLB export, web integration and mobile
performance. No acceptance of this study or expansion to 100 characters assumed.

## Reproduce

Use Blender in a separate background process, not the user's open scene:

```text
blender --background --factory-startup --python build_study.py
blender --background output/amani-3d-study.blend --python validate_study.py
blender --background output/amani-3d-study.blend --python render_motion.py
python pack_preview.py
```

The last step uses Pillow from the existing bundled workspace runtime.
Generated review frame PNGs and Blender backup files are not committed.
No changes to frontend/src, backend-go, package.json, StoryScene or prod.
