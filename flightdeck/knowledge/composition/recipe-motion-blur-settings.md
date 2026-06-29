# Recipe Comp Motion Blur Settings

Context: recipe `comp.motion_blur` support, proven with
`examples/recipes/minimal-comp-motion-blur.json`.

Recipe authoring:

- `shutter_angle` -> `SetShutterAngle`, integer degrees in `0..720`
- `shutter_phase` -> `SetShutterPhase`, integer raw phase value
- `adaptive_sample_limit` -> `SetMotionBlurAdaptiveSampleLimit`, non-negative integer
- `samples_per_frame` -> `SetMotionBlurSamplesPerFrame`, non-negative integer

Profile contract:

- `expected_profile.motion_blur.shutter_angle`
- `expected_profile.motion_blur.shutter_phase`
- `expected_profile.motion_blur.adaptive_sample_limit`
- `expected_profile.motion_blur.samples_per_frame`

Boundary:

- This slice covers the profile-visible shutter/sample settings.
- The composition motion blur enable flag is not exposed in the current stable
  profile schema and is not modeled in recipe IR here.
