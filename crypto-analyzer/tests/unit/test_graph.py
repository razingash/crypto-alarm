from core.analysis.graph import DependencyGraph


def test_add_formula():
    graph = DependencyGraph()

    graph.add_formula("a - b == 1000", 1)
    graph.add_formula("b + c == 100", 2)
    graph.add_formula("a - b + c <= 200", 3)
    assert len(graph.formulas) == 3
    assert len(graph.formula_ids) == 3

    # sympify почему-то возвращает False вместо адекватного значения
    assert str(graph.formulas[1]) == 'Eq(a - b, 1000)'
    assert str(graph.formulas[2]) == 'Eq(b + c, 100)'
    assert str(graph.formulas[3]) == 'a - b + c <= 200'

    assert "a - b" in graph.graph
    assert "b + c" in graph.graph
    assert "a - b + c" in graph.graph


def test_remove_formula():
    graph = DependencyGraph()

    graph.add_formula("a - b == 1000", 1)
    graph.add_formula("b + c == 100", 2)
    graph.add_formula("a - b + c <= 200", 3)

    graph.remove_formula(2)

    assert "Eq(b + c, 100)" not in graph.formula_ids
    assert 2 not in graph.formulas # 2 это id формулы
    assert "b + c" not in graph.graph # позже когда будет улушен граф этот случай должен будет ошибку вызывать


def test_remove_variable():
    graph = DependencyGraph()

    graph.add_formula("a + b == 1000", 142)
    graph.add_formula("a + b + c == 700", 51)
    graph.add_formula("a - c < 1000", 0)
    graph.add_formula("b + c <= 1300", 23)

    graph.update_variables_topological_Kahn({"a": 100, "b": 300, "c": 400})

    graph.remove_variable("a")

    assert 'a' not in graph.variables
    assert len(graph.variables) == 2
    assert len(graph.formulas) == 1


def test_get_formula_result_by_id():
    graph = DependencyGraph()

    graph.add_formula("a - b == 1000", 1)
    graph.update_variables_topological_Kahn({"a": 1200, "b": 200})

    result = graph.is_formula_triggered(1)
    assert result == True


def test_update_variables_topological_Kahn():
    graph = DependencyGraph()

    graph.add_formula("a - b == 1000", 1)
    graph.add_formula("b + c == 100", 2)
    graph.add_formula("a - b + c <= 200", 3)
    graph.add_formula("ETHBTC_priceChange < 1", 4)

    result = graph.update_variables_topological_Kahn({"a": 1200, "b": 200, "c": -100, 'ETHBTC_priceChange': -0.0003})

    assert graph.variables["a"] == 1200
    assert graph.variables["b"] == 200
    assert graph.variables["c"] == -100

    assert graph.is_formula_triggered(1) == True
    assert graph.is_formula_triggered(2) == True
    assert graph.is_formula_triggered(3) == False
    assert graph.is_formula_triggered(4) == True

    assert result == [1, 2, 4]

def test_get_formulas_variables():
    graph = DependencyGraph()

    graph.add_formula("ETHBTC_priceChange - ETHBTC_weightedAvgPrice == 1000", 1)
    graph.add_formula("ETHBTC_weightedAvgPrice + ETHBTC_lastPrice == 100", 2)
    graph.add_formula("ETHBTC_priceChange - ETHBTC_weightedAvgPrice + c <= 200", 3)

    formula_ids = graph.update_variables_topological_Kahn(
        {"ETHBTC_priceChange": 1200, "ETHBTC_weightedAvgPrice": 200, "ETHBTC_lastPrice": -100, "c": 122}
    )
    _, variable_values = graph.get_formulas_variables(formula_ids)

    assert variable_values == {'ETHBTC_weightedAvgPrice': 200, 'ETHBTC_priceChange': 1200, 'ETHBTC_lastPrice': -100}

def test_evaluate_variable_impact():
    graph = DependencyGraph()

    graph.add_formula("a - b == 1000", 1)
    graph.add_formula("b + c == 100", 2)
    graph.add_formula("a - b + c <= 200", 3)

    dependencies = graph.evaluate_variable_impact("a")
    assert len(dependencies) == 2
    dependencies = graph.evaluate_variable_impact("b")
    assert len(dependencies) == 3
    dependencies = graph.evaluate_variable_impact("c")
    assert len(dependencies) == 2

def test_evaluate_subexpression():
    graph = DependencyGraph()

    graph.add_formula("a - b == 1000", 1)
    graph.add_formula("b + c == 100", 2)
    graph.update_variables_topological_Kahn({"a": 1200, "b": 200, "c": -100})

    assert graph.evaluate_subexpression("a - b") == True
    assert graph.evaluate_subexpression("b + c") == True

def test_get_triggered_formulas():
    graph = DependencyGraph()

    graph.add_formula("a + b + c == 1000", 1)
    graph.add_formula("b + c == 100", 2)

    graph.update_variables_topological_Kahn({"a": 1200, "b": 200, "c": -100})
    result = graph.get_triggered_formulas([1, 2])

    assert 2 in result

def test_is_formula_triggered():
    graph = DependencyGraph()

    graph.add_formula("((a+b)/(abs(c)))<=1000", 1)
    graph.add_formula("((a + b) / c) <= 10000", 2)
    graph.update_variables_topological_Kahn({"a": 2.35927798, "b": -0.719, "c": -0.00017})

    assert graph.is_formula_triggered(1) == False
    assert graph.is_formula_triggered(2) == True
