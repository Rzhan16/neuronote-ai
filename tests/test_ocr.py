import pytest
from pathlib import Path
from ml.ocr_service import extract_layout

@pytest.fixture
def sample_image():
    """Create a sample test image if it doesn't exist."""
    image_path = Path("tests/resources/slide.jpg")
    if not image_path.exists():
        # Create a dummy image with some text
        from PIL import Image, ImageDraw, ImageFont
        img = Image.new('RGB', (800, 600), color='white')
        d = ImageDraw.Draw(img)
        
        # Add some text blocks
        texts = [
            ("Title", (100, 50)),
            ("Subtitle", (100, 150)),
            ("Bullet Point 1", (150, 250)),
            ("Bullet Point 2", (150, 300)),
            ("Footer", (100, 500))
        ]
        
        for text, pos in texts:
            d.text(pos, text, fill='black')
        
        img.save(image_path)
    
    return image_path

def test_extract_layout(sample_image):
    """Test that OCR extracts at least 5 text blocks."""
    with open(sample_image, 'rb') as f:
        image_bytes = f.read()
    
    result = extract_layout(image_bytes)
    
    assert 'blocks' in result
    assert len(result['blocks']) >= 5, f"Expected at least 5 blocks, got {len(result['blocks'])}"
    
    # Verify block structure
    for block in result['blocks']:
        assert 'text' in block
        assert 'bbox' in block
        assert len(block['bbox']) == 4  # x1, y1, x2, y2 