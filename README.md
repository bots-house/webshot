# 📸 [WebShot](https://webshot.bots.house)

Self-hosted web page screenshot generator.

## API

```http
GET https://webshot.bots.house/screenshot
```

| Param     | Type      | Description                                     | Default |
| :-------- | :-------- | :---------------------------------------------- | :------ |
| `width`   | `int`     | Viewport width in pixels of the browser render  | 1680    |
| `height`  | `int`     | Viewport height in pixels of the browser render | 867     |
| `scale`   | `float` | Viewport scale                                  | 1.0     |
| `format`  | `string`  | Output format (png, jpeg)                       | png     |
| `quality` | `int`     | Output image quiality                           | 100     |

