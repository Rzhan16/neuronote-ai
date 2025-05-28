from typing import Dict, List
import io
from PIL import Image
import pytesseract
import numpy as np

def extract_layout(file_bytes: bytes) -> Dict:
    """Extract text blocks and their bounding boxes from an image.
    
    Args:
        file_bytes: Raw bytes of the image file
        
    Returns:
        Dict with "blocks" key containing list of text and bbox
    """
    # Load image
    image = Image.open(io.BytesIO(file_bytes))
    
    # Convert to RGB if needed
    if image.mode != 'RGB':
        image = image.convert('RGB')
    
    # Limit to first 2 pages if PDF
    if hasattr(image, 'n_frames') and image.n_frames > 2:
        image = Image.open(io.BytesIO(file_bytes), pages=[0, 1])
    
    # Get document layout using Tesseract
    data = pytesseract.image_to_data(image, output_type=pytesseract.Output.DICT)
    
    # Extract text blocks and bounding boxes
    blocks = []
    width, height = image.size
    
    for i in range(len(data['text'])):
        if data['text'][i].strip():  # Only include non-empty text
            x = data['left'][i]
            y = data['top'][i]
            w = data['width'][i]
            h = data['height'][i]
            
            # Normalize coordinates to 0-1 range
            blocks.append({
                "text": data['text'][i],
                "bbox": [
                    x / width,
                    y / height,
                    (x + w) / width,
                    (y + h) / height
                ]
            })
    
    return {"blocks": blocks} 