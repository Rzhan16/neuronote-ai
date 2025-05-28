import pytest
from pathlib import Path
import wave
import numpy as np
from ml.asr_service import transcribe

@pytest.fixture
def sample_audio():
    """Create a sample test audio file if it doesn't exist."""
    audio_path = Path("tests/resources/sample.wav")
    if not audio_path.exists():
        audio_path.parent.mkdir(parents=True, exist_ok=True)
        
        # Create a 1-second sine wave audio
        sample_rate = 16000
        duration = 1  # seconds
        t = np.linspace(0, duration, int(sample_rate * duration))
        audio_data = np.sin(2 * np.pi * 440 * t)  # 440 Hz sine wave
        audio_data = (audio_data * 32767).astype(np.int16)
        
        with wave.open(str(audio_path), 'wb') as wav_file:
            wav_file.setnchannels(1)  # Mono
            wav_file.setsampwidth(2)  # 16-bit
            wav_file.setframerate(sample_rate)
            wav_file.writeframes(audio_data.tobytes())
    
    return audio_path

@pytest.fixture(autouse=True)
def mock_redis(mocker):
    """Mock Redis for all tests."""
    mock_redis = mocker.patch('ml.asr_service.redis_client')
    mock_redis.get.return_value = None
    return mock_redis

def test_transcribe(sample_audio, mock_redis):
    """Test that ASR returns a transcription."""
    # Read audio file as numpy array
    with wave.open(str(sample_audio), 'rb') as wav_file:
        sample_rate = wav_file.getframerate()
        n_frames = wav_file.getnframes()
        audio_data = wav_file.readframes(n_frames)
        audio_array = np.frombuffer(audio_data, dtype=np.int16).astype(np.float32) / 32768.0
    
    result = transcribe(audio_array)
    
    assert 'text' in result
    assert isinstance(result['text'], str)
    assert len(result['text']) > 0  # Should have some text, even if just noise

def test_transcribe_caching(sample_audio, mock_redis):
    """Test that transcriptions are cached in Redis."""
    # Read audio file as numpy array
    with wave.open(str(sample_audio), 'rb') as wav_file:
        sample_rate = wav_file.getframerate()
        n_frames = wav_file.getnframes()
        audio_data = wav_file.readframes(n_frames)
        audio_array = np.frombuffer(audio_data, dtype=np.int16).astype(np.float32) / 32768.0
    
    # First call should try cache, miss, and then set cache
    result1 = transcribe(audio_array)
    assert mock_redis.get.called
    assert mock_redis.set.called
    
    # Reset mock and simulate cache hit
    mock_redis.reset_mock()
    mock_redis.get.return_value = result1['text']
    
    # Second call should hit cache and return same result
    result2 = transcribe(audio_array)
    assert result2['text'] == result1['text']
    assert mock_redis.get.called
    assert not mock_redis.set.called  # Shouldn't set cache on hit 