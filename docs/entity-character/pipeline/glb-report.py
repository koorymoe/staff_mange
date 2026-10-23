import json, struct, sys, os

path = sys.argv[1]
with open(path, 'rb') as f:
    magic, ver, total = struct.unpack('<III', f.read(12))
    assert magic == 0x46546C67, 'not a glb'
    jlen, jtype = struct.unpack('<II', f.read(8))
    js = json.loads(f.read(jlen))

print('file: %s   %.2f MB' % (os.path.basename(path), os.path.getsize(path) / 1048576))
print('meshes      :', len(js.get('meshes', [])))
print('nodes       :', len(js.get('nodes', [])))
print('materials   :', len(js.get('materials', [])))
print('images      :', len(js.get('images', [])))
print('animations  :', len(js.get('animations', [])))
skins = js.get('skins', [])
print('skins       :', len(skins), '| joints:', [len(s.get('joints', [])) for s in skins])

# التعابير: أسماؤها بـ mesh.extras.targetNames
total_targets = 0
for m in js.get('meshes', []):
    names = (m.get('extras') or {}).get('targetNames')
    prims = m.get('primitives', [])
    ntar = len(prims[0].get('targets', [])) if prims and prims[0].get('targets') else 0
    if ntar:
        total_targets += ntar
        print('  shape keys on %-24s %3d' % (m.get('name', '?'), ntar), '(names present)' if names else '(no names!)')
print('TOTAL morph targets:', total_targets)

# أكبر الصور
imgs = js.get('images', [])
views = js.get('bufferViews', [])
sizes = []
for i, im in enumerate(imgs):
    bv = im.get('bufferView')
    n = views[bv]['byteLength'] if bv is not None else 0
    sizes.append((n, im.get('name', f'img{i}'), im.get('mimeType', '')))
sizes.sort(reverse=True)
print('--- أكبر ١٢ صورة ---')
for n, nm, mt in sizes[:12]:
    print('   %7.2f MB  %s  %s' % (n / 1048576, nm, mt))
print('مجموع الصور: %.2f MB' % (sum(n for n, _, _ in sizes) / 1048576))
