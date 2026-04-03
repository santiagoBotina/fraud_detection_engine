"""Fuzzy logic scorer for fraud detection.

Uses scikit-fuzzy to compute a fraud score (0-100) from transaction attributes.
Implements Requirements 2.1, 2.2, 2.3, 2.4.
"""

import hashlib
import math

import numpy as np
import skfuzzy as fuzz
from skfuzzy import control as ctrl


# Payment method risk mapping (Requirement 2.3)
PAYMENT_METHOD_RISK = {
    "BANK_TRANSFER": 1,
    "CARD": 3,
    "CRYPTO": 8,
}


def _ip_risk_score(ip: str) -> float:
    """Derive an IP risk score in [0, 10] from a hash-based heuristic.

    This is a deterministic mock for demo purposes: hash the IP string
    and map it to a value in [0, 10].
    """
    digest = hashlib.sha256(ip.encode("utf-8")).hexdigest()
    return int(digest, 16) % 11


class FuzzyLogicScorer:
    """Domain service that computes fraud scores using a fuzzy inference system.

    Input variables:
        - amount: transaction amount in cents (0 - 1,000,000)
        - payment_risk: risk derived from payment method (0 - 10)
        - ip_risk: risk derived from customer IP heuristic (0 - 10)

    Output variable:
        - fraud_score: integer in [0, 100]
    """

    def __init__(self) -> None:
        self._simulation = self._build_system()

    @staticmethod
    def _build_system() -> ctrl.ControlSystemSimulation:
        """Build the fuzzy control system with membership functions and rules."""

        # --- Input variables ---
        amount = ctrl.Antecedent(np.arange(0, 1_000_001, 1), "amount")
        payment_risk = ctrl.Antecedent(np.arange(0, 11, 1), "payment_risk")
        ip_risk = ctrl.Antecedent(np.arange(0, 11, 1), "ip_risk")

        # --- Output variable ---
        fraud_score = ctrl.Consequent(np.arange(0, 101, 1), "fraud_score")

        # --- Membership functions: amount ---
        amount["low"] = fuzz.trapmf(amount.universe, [0, 0, 5000, 20000])
        amount["medium"] = fuzz.trimf(amount.universe, [10000, 50000, 200000])
        amount["high"] = fuzz.trapmf(amount.universe, [100000, 500000, 1000000, 1000000])

        # --- Membership functions: payment_risk ---
        payment_risk["low"] = fuzz.trapmf(payment_risk.universe, [0, 0, 2, 4])
        payment_risk["medium"] = fuzz.trimf(payment_risk.universe, [2, 5, 8])
        payment_risk["high"] = fuzz.trapmf(payment_risk.universe, [6, 8, 10, 10])

        # --- Membership functions: ip_risk ---
        ip_risk["low"] = fuzz.trapmf(ip_risk.universe, [0, 0, 2, 4])
        ip_risk["medium"] = fuzz.trimf(ip_risk.universe, [2, 5, 8])
        ip_risk["high"] = fuzz.trapmf(ip_risk.universe, [6, 8, 10, 10])

        # --- Membership functions: fraud_score ---
        fraud_score["low"] = fuzz.trapmf(fraud_score.universe, [0, 0, 20, 40])
        fraud_score["medium"] = fuzz.trimf(fraud_score.universe, [30, 50, 70])
        fraud_score["high"] = fuzz.trapmf(fraud_score.universe, [60, 80, 100, 100])

        # --- Inference rules ---
        rule1 = ctrl.Rule(amount["low"] & payment_risk["low"], fraud_score["low"])
        rule2 = ctrl.Rule(amount["high"] & payment_risk["high"], fraud_score["high"])
        rule3 = ctrl.Rule(amount["high"] & ip_risk["high"], fraud_score["high"])
        rule4 = ctrl.Rule(payment_risk["high"] & ip_risk["high"], fraud_score["high"])
        rule5 = ctrl.Rule(amount["medium"] & payment_risk["medium"], fraud_score["medium"])
        rule6 = ctrl.Rule(amount["low"] & payment_risk["high"], fraud_score["medium"])
        rule7 = ctrl.Rule(
            amount["high"] & payment_risk["low"] & ip_risk["low"],
            fraud_score["medium"],
        )

        system = ctrl.ControlSystem([rule1, rule2, rule3, rule4, rule5, rule6, rule7])
        return ctrl.ControlSystemSimulation(system)

    def compute(self, amount_in_cents: int, payment_method: str, customer_ip: str) -> int:
        """Compute a fraud score for the given transaction attributes.

        Args:
            amount_in_cents: Transaction amount in cents (>= 0).
            payment_method: One of CARD, BANK_TRANSFER, CRYPTO.
            customer_ip: Customer IP address string.

        Returns:
            Integer fraud score in [0, 100].
        """
        # Clamp amount to the universe range
        clamped_amount = max(0, min(amount_in_cents, 1_000_000))

        # Map payment method to risk value, default to middle risk
        pay_risk = PAYMENT_METHOD_RISK.get(payment_method, 5)

        # Derive IP risk from heuristic
        ip_risk_val = _ip_risk_score(customer_ip)

        try:
            self._simulation.input["amount"] = clamped_amount
            self._simulation.input["payment_risk"] = pay_risk
            self._simulation.input["ip_risk"] = ip_risk_val
            self._simulation.compute()
            raw = self._simulation.output["fraud_score"]
        except Exception:
            # If fuzzy computation fails for any reason, default to 50 (medium)
            raw = 50.0

        # Handle NaN from defuzzification edge cases
        if raw is None or (isinstance(raw, float) and math.isnan(raw)):
            raw = 50.0

        # Round and clamp to [0, 100]
        score = int(round(raw))
        return max(0, min(score, 100))
