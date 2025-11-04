struct Particle {
    transform: mat3x3<f32>,
    velocity: vec2f,
}

@group(0) @binding(0)
var<storage, read_write> data: array<Particle>;

@group(0) @binding(1)
var<uniform> timestep: f32;

@compute @workgroup_size(1)
fn update_particles(
  @builtin(global_invocation_id) id: vec3<u32>
) {
  let i = id.x * 64 + id.y;

  let vel = vec3f(data[i].velocity * timestep, 1);
  let translate = mat3x3f(vec3f(1, 0, 0), vec3f(0, 1, 0), vel);

  // move the particle
  data[i].transform = translate * data[i].transform;
}
