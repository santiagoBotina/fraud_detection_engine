from dataclasses import dataclass


@dataclass(frozen=True)
class FraudSignalRequest:
    """Domain entity representing a fraud signal request from the FraudScore.Request Kafka topic.

    Maps to the FraudScore.Request Kafka message schema.
    """

    transaction_id: str
    amount_in_cents: int
    currency: str
    payment_method: str
    customer_id: str
    customer_ip_address: str
    timestamp: str

    def to_dict(self) -> dict:
        return {
            "transaction_id": self.transaction_id,
            "amount_in_cents": self.amount_in_cents,
            "currency": self.currency,
            "payment_method": self.payment_method,
            "customer_id": self.customer_id,
            "customer_ip_address": self.customer_ip_address,
            "timestamp": self.timestamp,
        }

    @classmethod
    def from_dict(cls, data: dict) -> "FraudSignalRequest":
        return cls(
            transaction_id=data["transaction_id"],
            amount_in_cents=data["amount_in_cents"],
            currency=data["currency"],
            payment_method=data["payment_method"],
            customer_id=data["customer_id"],
            customer_ip_address=data["customer_ip_address"],
            timestamp=data["timestamp"],
        )
