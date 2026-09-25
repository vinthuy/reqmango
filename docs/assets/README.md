# Assets for public docs

| File | Use |
|------|-----|
| `demo.gif` | README first-screen clip: Intake form → filled submit → Submitted → triage queue (accept/reject). ~230 KB. |
| `demo-0*.png` | Source frames used to build `demo.gif` (optional to keep). |

Rebuild GIF (Pillow):

```bash
python -c "from pathlib import Path; from PIL import Image; a=Path('docs/assets'); names=['demo-01-intake-form.png','demo-02-intake-filled.png','demo-03-intake-submitted.png','demo-04-triage-queue.png']; f=[Image.open(a/n).convert('RGBA') for n in names]; w,h=f[0].size; n=[];
for im in f:
 c=Image.new('RGBA',(w,h),(255,255,255,255)); s=min(w/im.width,h/im.height); nw,nh=int(im.width*s),int(im.height*s); r=im.resize((nw,nh),Image.Resampling.LANCZOS); c.paste(r,((w-nw)//2,(h-nh)//2),r); n.append(c.convert('P',palette=Image.Palette.ADAPTIVE,colors=128))
n[0].save(a/'demo.gif',save_all=True,append_images=n[1:],duration=[1600,1800,1400,2200],loop=0,optimize=True)"
```
