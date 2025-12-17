@group(0) @binding(0) var<storage, read_write> buf: array<u32>;

@group(0) @binding(1) var<uniform> seed: vec4u;

const width: u32 = 1024 * 2;
const height: u32 = 1024 * 2;

@compute
@workgroup_size(16, 1, 1)
fn compute(@builtin(global_invocation_id) id: vec3<u32>) {
    let idx = id.x;

    // particle coeff in real & imaginary space
    var cr = random_uniform(seed.x ^ idx) * 3.0 - 2.0;
    var ci = random_uniform(seed.y ^ idx) * 3.0 - 1.5;

    var path: array<vec2f, 4000>;

    // position of the particle;
    var zi: f32;
    var zr: f32;
    var n: u32;
    for (var i: u32 = 0; i < 400; i++) {
        let zr2 = zr*zr - zi*zi + cr;
        let zi2 = 2*zr*zi + ci;
        zr = zr2;
        zi = zi2;

        if zi*zi + zr*zr > 4.0 {
            // escaped, we can stop here
            n = i;

            if zr < -2.0 || zr > 1.5 || zi < -1.5 || zi > 1.5 {
                break;
            }
        }

        path[i] = vec2f(zr, zi);
    }

    for (var i: u32 = 0; i < n; i++) {
        let x = (path[i].x + 2.0) / 3.0;
        let y = (path[i].y + 1.5) / 3.0;

        let px = u32(x * f32(width) + 0.5);
        let py = u32(y * f32(height) + 0.5);

        if 0 <= px && px < width && 0 <= py && py < height {
            buf[py * width + px] += 1;
        }
    }
}

// A single iteration of Bob Jenkins' One-At-A-Time hashing algorithm for u32.
fn hash_u32(x_in: u32) -> u32 {
    var x = x_in;
    x += (x << 10u);
    x ^= (x >> 6u);
    x += (x << 3u);
    x ^= (x >> 11u);
    x += (x << 15u);
    return x;
}

// Construct a float with half-open range [0:1] using low 23 bits.
// All zeroes yields 0.0, all ones yields the next smallest representable value below 1.0.
fn float_construct_from_u32(m_in: u32) -> f32 {
    let ieeeMantissa = 0x007FFFFFu; // binary32 mantissa bitmask
    let ieeeOne = 0x3F800000u;      // 1.0 in IEEE binary32

    var m = m_in;
    m &= ieeeMantissa;              // Keep only mantissa bits (fractional part)
    m |= ieeeOne;                   // Add fractional part to 1.0

    let f = bitcast<f32>(m);        // Range [1:2]
    return f - 1.0;                 // Range [0:1]
}

// Pseudo-random value in half-open range [0:1] from a u32 seed.
fn random_uniform(seed: u32) -> f32 {
    return float_construct_from_u32(hash_u32(seed));
}
