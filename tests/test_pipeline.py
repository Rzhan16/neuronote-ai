import pytest
from fastapi import UploadFile
from unittest.mock import AsyncMock, MagicMock, patch
from ml.pipeline import process_note, detect_mimetype, extract_keyphrases

@pytest.fixture
def mock_db(mocker):
    """Mock database connection."""
    mock_conn = MagicMock()
    mock_cur = MagicMock()
    mock_conn.__enter__.return_value = mock_conn
    mock_conn.cursor.return_value.__enter__.return_value = mock_cur
    mocker.patch('ml.pipeline.get_db_connection', return_value=mock_conn)
    return mock_conn, mock_cur

@pytest.fixture
def mock_services(mocker):
    """Mock all ML services."""
    mocker.patch('ml.pipeline.extract_layout', return_value={
        'blocks': [{'text': 'Test text', 'bbox': [0, 0, 1, 1]}]
    })
    mocker.patch('ml.pipeline.transcribe', return_value={'text': 'Test transcript'})
    mocker.patch('ml.pipeline.summarise', return_value={'summary': 'Test summary'})
    mocker.patch('ml.pipeline.generate_qa', return_value=[
        {'q': 'Test question?', 'a': 'Test answer'}
    ])
    mocker.patch('ml.pipeline.keybert_model.extract_keywords', return_value=[
        ('key phrase', 0.8)
    ])

@pytest.fixture
def mock_file():
    """Create a mock file."""
    return AsyncMock(spec=UploadFile)

def test_detect_mimetype():
    """Test MIME type detection."""
    with patch('magic.from_buffer', return_value='image/jpeg'):
        assert detect_mimetype(b'test') == 'image/jpeg'

def test_extract_keyphrases():
    """Test keyphrase extraction."""
    with patch('ml.pipeline.keybert_model.extract_keywords', 
              return_value=[('key phrase', 0.8)]):
        result = extract_keyphrases('Test text')
        assert result == ['key phrase']

@pytest.mark.asyncio
async def test_process_note_image(mock_file, mock_services, mock_db):
    """Test processing an image note."""
    mock_file.read = AsyncMock(return_value=b'test image')
    with patch('ml.pipeline.detect_mimetype', return_value='image/jpeg'):
        result = await process_note(mock_file)
        
        assert 'note_id' in result
        mock_db[1].execute.assert_called()  # Check DB operations
        mock_db[0].commit.assert_called_once()

@pytest.mark.asyncio
async def test_process_note_audio(mock_file, mock_services, mock_db):
    """Test processing an audio note."""
    mock_file.read = AsyncMock(return_value=b'test audio')
    with patch('ml.pipeline.detect_mimetype', return_value='audio/wav'):
        result = await process_note(mock_file)
        
        assert 'note_id' in result
        mock_db[1].execute.assert_called()
        mock_db[0].commit.assert_called_once()

@pytest.mark.asyncio
async def test_process_note_invalid_type(mock_file):
    """Test processing an unsupported file type."""
    mock_file.read = AsyncMock(return_value=b'test data')
    with patch('ml.pipeline.detect_mimetype', return_value='application/pdf'):
        with pytest.raises(ValueError, match='Unsupported file type'):
            await process_note(mock_file) 