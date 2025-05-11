import random
import time

import pytest

from core.analysis.graph import DependencyGraph


def generate_random_formula(variables, num_variables):
    """
    в одной формуле максимум 100 переменных из-за того что sympy не такой мощный и примерно с 200 переменных начинаются
    ошибки из-за рекурсий(и это только простые выражения)
    """
    num_vars_in_formula = random.randint(3, min(num_variables, 100))
    selected_vars = random.sample(variables, num_vars_in_formula)
    ops = random.choices(["+", "-"], k=num_vars_in_formula - 1)
    expr = f" {random.choice(['==', '<', '>', '<=', '>='])} "
    formula = selected_vars[0]
    for i, var in enumerate(selected_vars[1:]):
        formula += f" {ops[i]} {var}"
    formula += expr + str(random.randint(10, 1000))
    return formula


@pytest.mark.performance
def test_add_and_update_formula_speed():
    graph = DependencyGraph()
    num_formulas = 3000
    num_variables = 3000
    variables = [f"var{i}" for i in range(1, num_variables + 1)]

    start_time = time.perf_counter()
    for i in range(num_formulas):
        formula = generate_random_formula(variables, num_variables)
        graph.add_formula(formula, i)
    add_time = time.perf_counter() - start_time

    print(f"{num_formulas} формул добавлено за {add_time:.2f} секунд.")

    updates = {var: random.randint(1, 1000) for var in variables}

    start_time = time.perf_counter()
    graph.update_variables_topological_Kahn(updates)
    update_time = time.perf_counter() - start_time

    print(f"Обновление переменных и пересчет занял {update_time:.2f} секунд.")

    assert update_time < 5, f"Обновление переменных заняло слишком много времени ({update_time} секунд)"
    assert add_time < 60, f"Добавление формул заняло слишком много времени ({add_time} секунд)"
