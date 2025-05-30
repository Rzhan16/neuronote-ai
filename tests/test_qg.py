import pytest
from ml.qg_service import generate_qa

@pytest.fixture(autouse=True)
def mock_redis(mocker):
    """Mock Redis for all tests."""
    mock_redis = mocker.patch('ml.qg_service.redis_client')
    mock_redis.get.return_value = None
    return mock_redis

def test_generate_qa(mock_redis):
    """Test question generation."""
    text = """
    The Python programming language was created by Guido van Rossum.
    It was first released in 1991 and has become very popular.
    Python is known for its simple syntax and readability.
    """
    
    result = generate_qa(text, max_q=3)
    
    assert isinstance(result, list)
    assert len(result) <= 3  # Should respect max_q
    for qa_pair in result:
        assert "q" in qa_pair
        assert "a" in qa_pair
        assert isinstance(qa_pair["q"], str)
        assert isinstance(qa_pair["a"], str)
        assert len(qa_pair["q"]) > 0
        assert len(qa_pair["a"]) > 0

def test_generate_qa_caching(mock_redis):
    """Test that QA pairs are cached in Redis."""
    text = "Python is a high-level programming language."
    
    # First call should try cache, miss, and then set cache
    result1 = generate_qa(text)
    assert mock_redis.get.called
    assert mock_redis.set.called
    
    # Reset mock and simulate cache hit
    mock_redis.reset_mock()
    import json
    mock_redis.get.return_value = json.dumps(result1)
    
    # Second call should hit cache and return same result
    result2 = generate_qa(text)
    assert result2 == result1
    assert mock_redis.get.called
    assert not mock_redis.set.called  # Shouldn't set cache on hit

def test_generate_qa_max_questions():
    """Test that max_questions parameter is respected."""
    text = """
    This is the first sentence.
    This is the second sentence.
    This is the third sentence.
    This is the fourth sentence.
    This is the fifth sentence.
    This is the sixth sentence.
    """
    
    result = generate_qa(text, max_q=4)
    assert len(result) <= 4  # Should not exceed max_q 