import os


class Config:
    KAFKA_BROKER_ADDRESS: str = os.getenv("KAFKA_BROKER_ADDRESS", "localhost:9092")
    KAFKA_CONSUMER_GROUP: str = os.getenv("KAFKA_CONSUMER_GROUP", "fraud-score-group")
    KAFKA_FRAUD_SIGNALS_REQUEST_TOPIC: str = os.getenv(
        "KAFKA_FRAUD_SIGNALS_REQUEST_TOPIC", "FraudSignals.Request"
    )
    KAFKA_FRAUD_SIGNALS_CALCULATED_TOPIC: str = os.getenv(
        "KAFKA_FRAUD_SIGNALS_CALCULATED_TOPIC", "FraudSignals.Calculated"
    )
    REDIS_HOST: str = os.getenv("REDIS_HOST", "localhost")
    REDIS_PORT: int = int(os.getenv("REDIS_PORT", "6379"))
    DYNAMO_DB_ENDPOINT: str = os.getenv("DYNAMO_DB_ENDPOINT", "http://localhost:8000")
    DYNAMO_DB_FRAUD_SCORES_TABLE: str = os.getenv(
        "DYNAMO_DB_FRAUD_SCORES_TABLE", "ddb-fraud-scores"
    )
    AWS_REGION: str = os.getenv("AWS_REGION", "us-east-1")
    AWS_ACCESS_KEY_ID: str = os.getenv("AWS_ACCESS_KEY_ID", "dummy")
    AWS_SECRET_ACCESS_KEY: str = os.getenv("AWS_SECRET_ACCESS_KEY", "dummy")
    QDRANT_HOST: str = os.getenv("QDRANT_HOST", "localhost")
    QDRANT_PORT: int = int(os.getenv("QDRANT_PORT", "6333"))


config = Config()
