#!/usr/bin/env python3
"""生成 AVProc 应用图标及功能图标"""

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


def create_preview_icon(size=256):
    """创建预览图标 - 播放按钮"""
    img = Image.new('RGBA', (size, size), (0, 0, 0, 0))
    draw = ImageDraw.Draw(img)
    
    margin = size // 10
    radius = size // 12
    draw.rounded_rectangle(
        [margin, margin, size-margin, size-margin],
        radius=radius,
        fill=(66, 133, 244)
    )
    
    # 播放三角形
    play_margin = size // 3
    center_x = size // 2 + size // 20
    center_y = size // 2
    
    h = size - 2 * play_margin
    w = int(h * 0.65)
    left = center_x - w // 2
    right = left + w
    top = center_y - h // 2
    bottom = center_y + h // 2
    
    points = [(left, top), (right, center_y), (left, bottom)]
    draw.polygon(points, fill=(255, 255, 255))
    
    return img


def create_convert_icon(size=256):
    """创建转换图标 - 双向箭头"""
    img = Image.new('RGBA', (size, size), (0, 0, 0, 0))
    draw = ImageDraw.Draw(img)
    
    margin = size // 10
    radius = size // 12
    draw.rounded_rectangle(
        [margin, margin, size-margin, size-margin],
        radius=radius,
        fill=(52, 168, 83)
    )
    
    # 双向箭头
    center = size // 2
    arrow_size = size // 4
    arrow_width = size // 8
    
    # 左箭头 (→)
    left_points = [
        (center - arrow_width, center - arrow_size),
        (center - arrow_width, center - arrow_size // 2),
        (center - arrow_width // 2, center - arrow_size // 2),
        (center - arrow_width // 2, center),
        (center - arrow_width, center),
        (center - arrow_width, center + arrow_size // 2),
        (center - arrow_width // 2, center + arrow_size // 2),
        (center - arrow_width // 2, center + arrow_size),
    ]
    
    # 右箭头 (→)
    right_points = [
        (center + arrow_width, center - arrow_size),
        (center + arrow_width, center - arrow_size // 2),
        (center + arrow_width // 2, center - arrow_size // 2),
        (center + arrow_width // 2, center),
        (center + arrow_width, center),
        (center + arrow_width, center + arrow_size // 2),
        (center + arrow_width // 2, center + arrow_size // 2),
        (center + arrow_width // 2, center + arrow_size),
    ]
    
    draw.polygon(left_points, fill=(255, 255, 255))
    draw.polygon(right_points, fill=(255, 255, 255))
    
    return img


def create_crop_icon(size=256):
    """创建剪裁图标 - 裁剪框"""
    img = Image.new('RGBA', (size, size), (0, 0, 0, 0))
    draw = ImageDraw.Draw(img)
    
    margin = size // 10
    radius = size // 12
    draw.rounded_rectangle(
        [margin, margin, size-margin, size-margin],
        radius=radius,
        fill=(251, 188, 5)
    )
    
    # 裁剪框 - 四个角
    inner = size // 4
    corner_len = size // 6
    c1 = size // 4
    c2 = size - size // 4
    
    # 左上角
    draw.line([(c1, c1), (c1 + corner_len, c1)], fill=(255, 255, 255), width=max(2, size // 32))
    draw.line([(c1, c1), (c1, c1 + corner_len)], fill=(255, 255, 255), width=max(2, size // 32))
    
    # 右上角
    draw.line([(c2, c1), (c2 - corner_len, c1)], fill=(255, 255, 255), width=max(2, size // 32))
    draw.line([(c2, c1), (c2, c1 + corner_len)], fill=(255, 255, 255), width=max(2, size // 32))
    
    # 左下角
    draw.line([(c1, c2), (c1 + corner_len, c2)], fill=(255, 255, 255), width=max(2, size // 32))
    draw.line([(c1, c2), (c1, c2 - corner_len)], fill=(255, 255, 255), width=max(2, size // 32))
    
    # 右下角
    draw.line([(c2, c2), (c2 - corner_len, c2)], fill=(255, 255, 255), width=max(2, size // 32))
    draw.line([(c2, c2), (c2, c2 - corner_len)], fill=(255, 255, 255), width=max(2, size // 32))
    
    return img


def create_merge_icon(size=256):
    """创建拼接图标 - 连接符号"""
    img = Image.new('RGBA', (size, size), (0, 0, 0, 0))
    draw = ImageDraw.Draw(img)
    
    margin = size // 10
    radius = size // 12
    draw.rounded_rectangle(
        [margin, margin, size-margin, size-margin],
        radius=radius,
        fill=(234, 67, 53)
    )
    
    # 两个矩形连接
    gap = size // 8
    box_w = (size - 2 * margin - gap) // 2
    box_h = size // 3
    box_y = (size - box_h) // 2
    
    # 左矩形
    left_x = margin + gap // 2
    draw.rounded_rectangle(
        [left_x, box_y, left_x + box_w, box_y + box_h],
        radius=max(2, size // 24),
        fill=(255, 255, 255)
    )
    
    # 右矩形
    right_x = left_x + box_w + gap
    draw.rounded_rectangle(
        [right_x, box_y, right_x + box_w, box_y + box_h],
        radius=max(2, size // 24),
        fill=(255, 255, 255)
    )
    
    # 中间加号
    plus_len = size // 8
    plus_thick = max(2, size // 16)
    cx = size // 2
    cy = size // 2
    draw.line([(cx - plus_len, cy), (cx + plus_len, cy)], fill=(255, 255, 255), width=plus_thick)
    draw.line([(cx, cy - plus_len), (cx, cy + plus_len)], fill=(255, 255, 255), width=plus_thick)
    
    return img


def save_icon(img, name, sizes):
    """保存图标到多个尺寸"""
    for size in sizes:
        resized = img.resize((size, size), Image.LANCZOS)
        resized.save(f'{name}_{size}.png')
    img.save(f'{name}.png')
    print(f'已生成 {name}.png 及各尺寸版本')


def save_ico(img, name):
    """保存 ICO 文件"""
    ico_sizes = [(16, 16), (32, 32), (48, 48), (256, 256)]
    ico_images = []
    for size in ico_sizes:
        ico_images.append(img.resize(size, Image.LANCZOS))
    ico_images[0].save(f'{name}.ico', format='ICO', sizes=ico_sizes)
    print(f'已生成 {name}.ico')


def main():
    sizes = [16, 32, 48, 64, 128, 256]
    
    # 生成主应用图标
    print("=== 生成主应用图标 ===")
    img = create_icon(256)
    for size in sizes:
        resized = img.resize((size, size), Image.LANCZOS)
        resized.save(f'icon_{size}.png')
    img.save('icon.png')
    save_ico(img, 'icon')
    print('已生成 icon.png')
    
    # 生成预览图标
    print("\n=== 生成预览图标 ===")
    img_preview = create_preview_icon(256)
    save_icon(img_preview, 'preview', sizes)
    
    # 生成转换图标
    print("\n=== 生成转换图标 ===")
    img_convert = create_convert_icon(256)
    save_icon(img_convert, 'convert', sizes)
    
    # 生成剪裁图标
    print("\n=== 生成剪裁图标 ===")
    img_crop = create_crop_icon(256)
    save_icon(img_crop, 'crop', sizes)
    
    # 生成拼接图标
    print("\n=== 生成拼接图标 ===")
    img_merge = create_merge_icon(256)
    save_icon(img_merge, 'merge', sizes)
    
    print("\n=== 完成 ===")


if __name__ == '__main__':
    main()
