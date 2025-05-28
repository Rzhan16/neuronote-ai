import pytest
from ml.summarise_service import summarise

@pytest.fixture(autouse=True)
def mock_redis(mocker):
    """Mock Redis for all tests."""
    mock_redis = mocker.patch('ml.summarise_service.redis_client')
    mock_redis.get.return_value = None
    return mock_redis

def test_summarise_paragraph(mock_redis):
    """Test paragraph-style summarization."""
    text = """
    The researchers conducted a comprehensive study on climate change impacts. 
    They analyzed data from multiple sources spanning several decades. 
    The results showed significant temperature increases in urban areas. 
    Additionally, they found correlations between rising temperatures and changing precipitation patterns. 
    The study concluded with recommendations for policy makers and urban planners.
    """
    
    result = summarise(text, style="paragraph")
    
    assert "summary" in result
    assert isinstance(result["summary"], str)
    assert len(result["summary"]) > 0
    assert len(result["summary"]) < len(text)  # Summary should be shorter than input

def test_summarise_bullets(mock_redis):
    """Test bullet-point style summarization."""
    text = """
    The researchers conducted a comprehensive study on climate change impacts. 
    They analyzed data from multiple sources spanning several decades. 
    The results showed significant temperature increases in urban areas. 
    Additionally, they found correlations between rising temperatures and changing precipitation patterns. 
    The study concluded with recommendations for policy makers and urban planners.
    """
    
    result = summarise(text, style="bullets")
    
    assert "summary" in result
    assert isinstance(result["summary"], str)
    assert len(result["summary"]) > 0
    assert "•" in result["summary"]  # Should contain bullet points

def test_summarise_caching(mock_redis):
    """Test that summaries are cached in Redis."""
    text = "This is a test text that needs to be summarized."
    
    # First call should try cache, miss, and then set cache
    result1 = summarise(text)
    assert mock_redis.get.called
    assert mock_redis.set.called
    
    # Reset mock and simulate cache hit
    mock_redis.reset_mock()
    mock_redis.get.return_value = result1["summary"]
    
    # Second call should hit cache and return same result
    result2 = summarise(text)
    assert result2["summary"] == result1["summary"]
    assert mock_redis.get.called
    assert not mock_redis.set.called  # Shouldn't set cache on hit 