# 📸 [WebShot](https://webshot.bots.house)

Self-hosted web page screenshot generator.

## API

```http
GET https://webshot.bots.house/screenshot
```

| Param         |   Type    | Description                                             |   Default    |
| :------------ | :-------: | :------------------------------------------------------ | :----------: |
| `url`         | `string`  | URL of target page                                      | **Required** |
| `width`       |   `int`   | Viewport width in pixels of the browser render          |     1680     |
| `height`      |   `int`   | Viewport height in pixels of the browser render         |     867      |
| `scale`       |  `float`  | Viewport scale                                          |     1.0      |
| `format`      | `string`  | Output format (png, jpeg)                               |     png      |
| `quality`     |   `int`   | Output image quiality                                   |     100      |
| `clip_x`      | `float64` | X offset in device independent pixels (dip).            |     null     |
| `clip_y`      | `float64` | Y offset in device independent pixels (dip).            |     null     |
| `clip_width`  | `float64` | Rectangle width in device independent pixels (dip).     |     null     |
| `clip_height` | `float64` | Rectangle height in device independent pixels (dip).    |     null     |
| `delay`       |   `int`   | Delay in milliseconds, to wait after the page is loaded |     null     |
