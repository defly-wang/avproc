#!/usr/bin/env python3
"""生成 AVProc 应用图标"""

from PIL import Image, ImageDraw

def create_icon(size=256):
    """创建一个简单的视频图标"""
    img = Image.new('RGBA', (size, size), (0, 0, 0, 0))
    draw = ImageDraw.Draw(img)
    
    # 背景圆角矩形
    margin = size // 16
    radius = size // 8
    draw.rounded_rectangle(
        [margin, margin, size-margin, size-margin],
        radius=radius,
        fill=(66, 133, 244)
    )
    
    # 播放按钮 (三角形)
    play_margin = size // 4
    center_x = size // 2 + size // 16
    center_y = size // 2
    
    # 计算三角形顶点
    h = size - 2 * play_margin
    w = int(h * 0.6)
    left = center_x - w // 2
    right = left + w
    top = center_y - h // 2
    bottom = center_y + h // 2
    
    points = [
        (left, top),
        (right, center_y),
        (left, bottom)
    ]
    draw.polygon(points, fill=(255, 255, 255))
    
    return img

def main():
    sizes = [16, 32, 48, 64, 128, 256]
    
    # 生成 PNG 图标
    for size in sizes:
        img = create_icon(size)
        img.save(f'icon_{size}.png')
        print(f'已生成 icon_{size}.png')
    
    # 生成 256x256 作为主图标
    img_256 = create_icon(256)
    img_256.save('icon.png')
    print('已生成 icon.png')
    
    # 生成 ICO (Windows)
    ico_sizes = [(16, 16), (32, 32), (48, 48), (256, 256)]
    ico_images = []
    for size in ico_sizes:
        ico_images.append(create_icon(size[0]).resize(size))
    ico_images[0].save('icon.ico', format='ICO', sizes=ico_sizes)
    print('已生成 icon.ico')

if __name__ == '__main__':
    main()
