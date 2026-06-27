from decimal import Decimal, ROUND_HALF_UP


class ValidationError(ValueError):
    pass


DEFAULT_DISCOUNT_RULES = {
    "VIP": Decimal("0.20"),
}

MAX_ITEMS = 100


def _money(value):
    try:
        amount = Decimal(str(value))
    except Exception as exc:
        raise ValidationError("价格必须是可解析的数字") from exc

    if amount < 0:
        raise ValidationError("价格不能为负数")

    return amount


def _quantity(value):
    if not isinstance(value, int):
        raise ValidationError("数量必须是整数")
    if value <= 0:
        raise ValidationError("数量必须大于 0")
    return value


def calculate_checkout(cart, request_id, discount_rules=None):
    if not request_id:
        raise ValidationError("request_id 不能为空，用于幂等和审计追踪")

    if not isinstance(cart, dict):
        raise ValidationError("cart 必须是字典")

    items = cart.get("items")
    if not items:
        raise ValidationError("购物车不能为空")
    if len(items) > MAX_ITEMS:
        raise ValidationError("购物车商品数量超过上限")

    rules = discount_rules or DEFAULT_DISCOUNT_RULES
    subtotal = Decimal("0")

    for index, item in enumerate(items):
        if not isinstance(item, dict):
            raise ValidationError(f"第 {index + 1} 个商品格式错误")
        price = _money(item.get("price"))
        qty = _quantity(item.get("qty", 1))
        subtotal += price * qty

    coupon = cart.get("coupon")
    discount_rate = rules.get(coupon, Decimal("0"))
    discount = (subtotal * discount_rate).quantize(Decimal("0.01"), rounding=ROUND_HALF_UP)
    total = (subtotal - discount).quantize(Decimal("0.01"), rounding=ROUND_HALF_UP)

    return {
        "status": "ok",
        "request_id": request_id,
        "subtotal": str(subtotal.quantize(Decimal("0.01"), rounding=ROUND_HALF_UP)),
        "discount": str(discount),
        "total": str(total),
        "currency": "CNY",
        "audit_events": [
            f"checkout.calculated request_id={request_id}",
            f"items={len(items)} coupon={coupon or '-'} discount_rate={discount_rate}",
        ],
    }
