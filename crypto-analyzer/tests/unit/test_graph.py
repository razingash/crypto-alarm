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

    graph.update_variables_topological_Kahn({"a": 1200, "b": 200, "c": -100})

    assert graph.variables["a"] == 1200
    assert graph.variables["b"] == 200
    assert graph.variables["c"] == -100

    assert graph.is_formula_triggered(1) == True
    assert graph.is_formula_triggered(2) == True
    assert graph.is_formula_triggered(3) == False


def test_evaluate_variable_impact():  # не работает правильно метод в классе
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

    result = int(graph.evaluate_subexpression("a - b"))
    assert result == 1000

    result = int(graph.evaluate_subexpression("b + c"))
    assert result == 100

    result1 = graph.evaluate_subexpression("a - b")
    result2 = graph.evaluate_subexpression("a - b")
    assert result1 == result2

"""
def test_evaluate_formula():
    graph = DependencyGraph()

    graph.add_formula("a + b + c == 1000")
    graph.add_formula("b + c == 100")

    graph.update_variables_topological_Kahn({"a": 1200, "b": 200, "c": -100})

    result = graph.evaluate_formula(0)
    print(result)
    #assert result == 1000

    #result = graph.evaluate_formula(1)
    #assert result == 100

    #result = graph.evaluate_formula(2)
    #assert result == 1200
"""