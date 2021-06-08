# 📸 [WebShot](https://webshot.bots.house)

Self-hosted web page screenshot generator.

## API

```http
GET https://webshot.bots.house/image
```

| Param         |   Type    | Description                                                   |   Default    |
| :------------ | :-------: | :------------------------------------------------------------ | :----------: |
| `url`         | `string`  | URL of target page                                            | **Required** |
| `width`       |   `int`   | Viewport width in pixels of the browser render                |     1680     |
| `height`      |   `int`   | Viewport height in pixels of the browser render               |     867      |
| `scale`       |  `float`  | Viewport scale                                                |     1.0      |
| `format`      | `string`  | Output format (png, jpeg)                                     |     png      |
| `quality`     |   `int`   | Output image quiality                                         |     100      |
| `clip_x`      | `float64` | X offset in device independent pixels (dip).                  |     null     |
| `clip_y`      | `float64` | Y offset in device independent pixels (dip).                  |     null     |
| `clip_width`  | `float64` | Rectangle width in device independent pixels (dip).           |     null     |
| `clip_height` | `float64` | Rectangle height in device independent pixels (dip).          |     null     |
| `delay`       |   `int`   | Delay in milliseconds, to wait after the page is loaded       |     null     |
| `full_page`   |  `bool`   | Capture full page screenshot                                  |    false     |
| `scroll_page` |  `bool`   | Scroll through the entire page before capturing a screenshot. |    false     |

## Deploy

### Heroku

[![Deploy Heroku](https://www.herokucdn.com/deploy/button.svg)](https://heroku.com/deploy?template=https://github.com/bots-house/webshot/tree/main)


### DigitalOcean

[![Deploy to DigitalOcean](https://www.deploytodo.com/do-btn-blue.svg)](https://cloud.digitalocean.com/apps/new?repo=https://github.com/bots-house/webshot/tree/main)