import bpy
from collections import Counter
bpy.ops.wm.open_mainfile(filepath="amani-studio-review.blend")
mesh=next(o for o in bpy.data.objects if o.type=='MESH')
c=Counter()
over=0; lost=0.0; worst=0.0
for v in mesh.data.vertices:
    gs=sorted([g.weight for g in v.groups if g.weight>1e-6], reverse=True)
    c[len(gs)]+=1
    if len(gs)>4:
        over+=1
        tail=sum(gs[4:]); tot=sum(gs)
        if tot>0:
            lost+=tail/tot
            worst=max(worst,tail/tot)
print("توزيع عدد التأثيرات لكل رأس:", dict(sorted(c.items())))
print(f"رؤوس فوق ٤ تأثيرات: {over} من {len(mesh.data.vertices)} = {over/len(mesh.data.vertices)*100:.2f}٪")
if over: print(f"متوسط الوزن المفقود بهالرؤوس: {lost/over*100:.2f}٪ | أسوأ رأس: {worst*100:.2f}٪")
