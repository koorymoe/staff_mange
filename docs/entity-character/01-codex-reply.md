# (M) Engineering recommendation: choose SVG (A)

## Decision

Choose **SVG (A)**. The product requirement is a living, expressive companion from the first second; it is not a requirement for a 3D engine. SVG directly supplies independent pupils, brows, mouth, and head layers, with deterministic colours and zero runtime asset generation. This is the smallest reliable path to a character that feels present rather than a swapped image.

Babylon is already available, but a new render loop, canvas ownership, resize handling, context-loss handling, and weak-device validation are extra product risk. We should not trade the core interaction for unmeasured 3D polish at this stage.

## Visual evidence

`proto/entity-svg-prototype.html` is a standalone, inspectable prototype. It contains only a head, two eye whites, pupils, and angry brows. Move the pointer over it: the pupils move independently while the head stays still; hover also changes brow tension. It demonstrates the exact interaction gap in the current PNG approach without a library, generated image, or external call.

The production avatar must expose the agreed `EntityAvatarProps` contract. The rig, not the avatar, will provide `lookX`, `lookY`, blink, talk, breathe, reduced-motion behaviour, and visibility pause.

## Performance position

No Babylon route is selected, so there is no new engine bundle or canvas render loop to measure. The prototype is inline SVG plus event-driven transforms; it has no timer or animation frame loop. Before implementation, we should still record a baseline and post-change `npm run build` output rather than predict a byte count.

## Risks in the selected SVG route

1. **Generic/flat appearance.** Code-only vector art can look like a dashboard icon instead of a companion. Mitigation: define one compact visual language (silhouette, palette, proportions, expression range) before adding variants.
2. **Expression coupling.** Several independently transformed layers can produce awkward expressions if their ranges are not bounded. Mitigation: centralize clamps and named pose presets in the rig; the SVG receives only normalized props.
3. **Accessibility and motion fatigue.** Constant gaze/breathing can distract users. Mitigation: honour `prefers-reduced-motion`, stop the rig when hidden, and cap movement amplitudes; SVG itself stays pure and has no loop.

## Exit criterion for this choice

Approve SVG only if the first integrated version proves all three: pupil movement is visibly independent of the head, mood changes are pose transitions rather than image swaps, and the existing briefing/drag/chat behaviours remain unchanged. Otherwise reopen the 3D decision with measured evidence.
