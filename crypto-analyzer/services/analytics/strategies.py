

"""тут тоже падла надо будет заморочится чтобы было конкурентно из-за того
что если перебирать все имеющиеся стратегии то это будет кошмар, слишком большая нагрузка

сделать просто логику стратегий, приоритеты будут касатся только стратегий пользователей
"""

def strategy_price_jump(is_percentage: bool, min_leap: float, max_leap: float, current_price: float, fixed_price: float):
    """пробитие предела цены - может быть процентным и числовым"""
    if is_percentage:
        lower_bound = current_price * (1 - fixed_price / 100)
        upper_bound = current_price * (1 + fixed_price / 100)
    else:
        lower_bound = min_leap
        upper_bound = max_leap

    # изменить на коды ответов?
    if current_price < lower_bound:
        return "нижний предел пробит"
    if current_price > upper_bound:
        return "верхний предел пробит"

    return None
