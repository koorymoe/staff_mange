"""Pack Blender-rendered frames. No AI image editing or new artwork."""
from pathlib import Path
from PIL import Image
out=Path(__file__).resolve().parent/'output'
frames=[Image.open(p).convert('RGB') for p in sorted((out/'motion-frames').glob('*.png'))]
assert len(frames)==60, f'Expected 60 review frames, got {len(frames)}'
frames[0].save(out/'turn-and-walk.gif',save_all=True,append_images=frames[1:],duration=125,loop=0,optimize=False)
print(out/'turn-and-walk.gif')
