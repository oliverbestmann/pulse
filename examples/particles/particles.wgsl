struct Particle {
    transform: mat3x3<f32>,
    color: vec4f,
    velocity: vec2f,
}

struct SpriteInstance {
    color: array<f32, 4>,

    tr_row0: array<f32, 3>,
    tr_row1: array<f32, 3>,

    source_rect: Rect,
    target_rect: Rect,
};

struct Rect {
    pos_packed: u32,
    size_packed: u32,
}

fn pack_rect(x: u32, y: u32, w: u32, h: u32) -> Rect {
    let pos = (y << 16) | (x & 0xffff);
    let size = (h << 16) | (w & 0xffff);

    var rect: Rect;
    rect.pos_packed = pos;
    rect.size_packed = size;
    return rect;
}

@group(0) @binding(0)
var<storage, read_write> data: array<Particle>;

@group(0) @binding(1)
var<storage, read_write> sprites: array<SpriteInstance>;

@group(0) @binding(2)
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

  // write to output buffer
  sprites[i].color = array(
    data[i].color.r,
    data[i].color.g,
    data[i].color.b,
    data[i].color.a,
  );

  sprites[i].source_rect = pack_rect(0, 0, 64, 64);
  sprites[i].target_rect = pack_rect(0, 0, 0xffff, 0xffff);

  let tr = transpose(data[i].transform);
  sprites[i].tr_row0[0] = tr[0].x;
  sprites[i].tr_row0[1] = tr[0].y;
  sprites[i].tr_row0[2] = tr[0].z;

  sprites[i].tr_row1[0] = tr[1].x;
  sprites[i].tr_row1[1] = tr[1].y;
  sprites[i].tr_row1[2] = tr[1].z;
}
