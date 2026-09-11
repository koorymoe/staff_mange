import json,struct,sys
def load(p):
    d=open(p,'rb').read(); assert d[:4]==b'glTF'
    o=12; js=None
    while o < len(d):
        ln,ty = struct.unpack_from('<II', d, o); o+=8
        if ty==0x4E4F534A: js=json.loads(d[o:o+ln])
        o+=ln
    return js, len(d)
for p in sys.argv[1:]:
    js,total = load(p)
    bv=js['bufferViews']
    def sz(view_ids): return sum(bv[i]['byteLength'] for i in set(view_ids))
    acc=js['accessors']
    print(f"\n=== {p} — {total/1048576:.2f} م.ب ===")
    print("مقاطع:", len(js.get('animations',[])))
    anim_total=0
    rows=[]
    for a in js.get('animations',[]):
        ids=[]
        for s in a['samplers']:
            for k in ('input','output'):
                v=acc[s[k]].get('bufferView')
                if v is not None: ids.append(v)
        b=sz(ids); anim_total+=b
        rows.append((b,a.get('name','?')))
    for b,n in sorted(rows, reverse=True): print(f"   {n:16s} {b/1024:7.0f} ك.ب")
    imgs=sum(bv[i['bufferView']]['byteLength'] for i in js.get('images',[]) if 'bufferView' in i)
    print(f"  حركات={anim_total/1024:.0f} ك.ب | صور={imgs/1024:.0f} ك.ب | صور عدد={len(js.get('images',[]))}")
    print("  عُقد:", len(js['nodes']), "| جلود:", len(js.get('skins',[])), "| مفاصل:", len(js['skins'][0]['joints']) if js.get('skins') else 0)
    n=[x for x in js['nodes'] if 'scale' in x]
    print("  عقد فيها مقياس:", [(x.get('name'),[round(v,4) for v in x['scale']]) for x in n][:3])
