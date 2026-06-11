# Screenshots

Drop your captures in this folder so they render in the root `README.md` and on the GitHub project page.

## Recommended captures

| File | What to show |
| --- | --- |
| `dashboard.png` | The full dashboard with a few active downloads and the 3D backdrop. |
| `add.png` | The "Pounce" add bar with the Options panel open (connections, speed, yt-dlp). |
| `themes.png` | The theme switcher, ideally a side-by-side or the Settings modal. |
| `demo.gif` | A short screen recording of a download running + a theme switch. |

## Tips

- Capture at 2× (retina) for crisp images, then export PNG.
- Keep GIFs under ~8 MB so they load fast in the README.
- Use a dark desktop background — it blends with Pounce's glassmorphism.

## Recording `demo.gif` (one recipe)

1. Record your screen to `demo.mp4` with any tool (OBS, macOS `Cmd+Shift+5`, Windows Game Bar, GNOME screen recorder). Show: paste a URL → download accelerates → pause/resume → switch a theme.
2. Convert to a crisp, small GIF using a generated palette:

```bash
ffmpeg -i demo.mp4 -vf "fps=15,scale=1000:-1:flags=lanczos,palettegen" /tmp/palette.png
ffmpeg -i demo.mp4 -i /tmp/palette.png \
  -filter_complex "fps=15,scale=1000:-1:flags=lanczos[x];[x][1:v]paletteuse" \
  docs/screenshots/demo.gif
```

Aim for < 8 MB. Then reference it in the root `README.md` Demo section.

_These files are intentionally git-ignored-friendly: commit only the final, optimized images._
