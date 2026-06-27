def calculate_checkout(cart):
    total = sum(item["price"] * item.get("qty", 1) for item in cart["items"])

    if cart.get("coupon") == "VIP":
        total *= 0.8

    return {"status": "ok", "total": round(total, 2)}
