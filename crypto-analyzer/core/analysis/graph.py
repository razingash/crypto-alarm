import hashlib
from collections import defaultdict, deque

from sympy import sympify, lambdify, Mul, Add, preorder_traversal, Symbol

from core.logger import custom_logger

"""
!!!
изменить evaluate_subexpression и evaluate_formula потому что sympify().subs() возвращает разные типы в зависимости от того
как отработает
"""
class DependencyGraph:
    """
    Горячее хранилище активных выражений; хранит выражения в виде графа зависимостей
    1) улучшить добавление(можно позже, чтобы учитывало числа -> скобки -> циклические выражения(самое трудное) ->)
    2) оптимизировать добавление (можно позже)
    """

    def __init__(self):
        self.graph = defaultdict(set)  # {переменная -> набор формул, которые от неё зависят}
        self.formulas = {}  # {ID формулы -> символьное выражение}
        self.compiled = {}  # {ID формулы -> скомпилированная функция}
        self.variables = {}  # {переменная -> текущее значение}
        self.cache = {}  # Кэш для промежуточных результатов
        self.formula_ids = {}  # Formula: ID | нужно чтобы нормально удалять их
        self.subexpression_weights = defaultdict(int) # {подвыражение -> количество повторений}

    def add_formula(self, formula: str, formula_id: int): # слишком медленная
        """Добавляет формулу в граф зависимостей"""
        try:
            expr = sympify(formula, evaluate=False)
            if not hasattr(expr, "free_symbols"):
                raise ValueError(f"Incorrect expression: {formula}")

            self.formulas[formula_id] = expr
            self.formula_ids[str(expr)] = formula_id

            # разбивка на подвыражения
            for subexpr in preorder_traversal(expr):
                if subexpr.is_Atom: # numbers and expressions are excluded
                    continue
                self.graph[str(subexpr)].add(formula_id)
                self.subexpression_weights[str(subexpr)] += 1

            variables = list(expr.free_symbols)
            self.compiled[formula_id] = lambdify(variables, expr, "numpy")
        except RuntimeError as e: # возможно добавить больше конкретики
            custom_logger.log_with_path(
                level=2,
                msg=f"Recursion error due to a too heavy formula, it is necessary to either increase the recursion"
                    f"limit or reduce the permissible severity of the formulas:  {formula} \nError\n {e}",
                filename="Graph.log"
            )
            print('опасная формула')
            return 'reccursion error'
        except Exception as e:
            print(f"Ошибка в формуле '{formula}': {e}")
            return f"Ошибка в формуле '{formula}': {e}"
        print('success', self.__dict__)
        return True

    def remove_formula(self, formula_id: int):
        """Удаляет формулу из графа и все связанные подвыражения"""
        try:
            if formula_id not in self.formulas:
                custom_logger.log_with_path(
                    level=2,
                    msg=f"The formula was not found, most likely due to a bug in the code, or asynchronous approach: {formula_id}",
                    filename="Graph.log"
                )
                print(f"Формула с id '{formula_id}' не найдена.")
                return False

            expr = self.formulas[formula_id]
            del self.formulas[formula_id]
            del self.compiled[formula_id]
            self.formula_ids.pop(str(expr), None)

            for subexpr, formula_ids in list(self.graph.items()):
                if formula_id in formula_ids:
                    self.graph[subexpr].remove(formula_id)
                    if not self.graph[subexpr]:
                        del self.graph[subexpr]

            # удаление подвыражений, если не используются
            for subexpr in list(self.subexpression_weights.keys()):
                if subexpr in self.graph:
                    continue
                del self.subexpression_weights[subexpr]

            print(f"Удалена формула: {expr}")
        except Exception as e:
            print(f"Ошибка при удалении формулы '{formula_id}': {e}")
            custom_logger.log_with_path(
                level=1,
                msg=f"Formula with id {formula_id} wasn't removed from graph due to an error: {e}",
                filename="Graph.log"
            )
            return f"Ошибка при удалении формулы '{formula_id}': {e}"
        print('deleted, ', self.__dict__)
        return True

    def remove_variable(self, variable_str: str) -> list:
        """
        удаляет переменную и все подвязанные формулы и подвыражения\n
        используется в случае если на Binance уберут какой-нибудь параметр или валюту\n
        возвращает список ID удаленных формул
        """
        try:
            variable = sympify(variable_str)
            if variable_str not in self.variables:
                print(f"Переменная '{variable_str}' не найдена.")
                return []

            # выборка связанных формул
            dependent_formulas = set()
            for formula_id, formula in self.formulas.items():
                if variable in formula.free_symbols:
                    dependent_formulas.add(formula_id)

            if not dependent_formulas:
                print(f"Нет формул, зависящих от переменной '{variable_str}'.")
                return []

            removed_formula_ids = list(dependent_formulas)

            # удаление подвязаных формул
            for formula_id in dependent_formulas:
                self.remove_formula(formula_id)

            del self.variables[variable_str]

            # удаление подвыражений
            for subexpr in list(self.subexpression_weights.keys()):
                if variable_str in subexpr:
                    del self.subexpression_weights[subexpr]

            print(f"Переменная '{variable_str}' и все связанные формулы и подвыражения удалены.")
            return removed_formula_ids

        except Exception as e:
            print(f"Ошибка при удалении переменной '{variable_str}': {e}")
            return []

    def update_variables_topological_Kahn(self, updates: dict) -> list:
        """
        алгоритм Кана (не учитывает циклические зависимости)
        Обновляет сразу несколько переменных и пересчитывает только необходимые формулы

        Скорость выполнения зависит от количества уникальных параметров:
        Для 3000 уникальных параметров и 10000 разных формул с количеством пременных равным 3000 обновление
        3000 параметров(по сути всего графа) происхоидит примерно за 0.0004 секунды
        """
        print('Updating variables with data:', updates)

        self.variables.update(updates)

        affected_formulas = set()
        for var in updates:
            deps = self.graph.get(var)
            if deps:
                affected_formulas.update(deps)
            else:
                for f_id, expr in self.formulas.items():
                    if any(str(sym) == var for sym in expr.free_symbols):
                        affected_formulas.add(f_id)

        if not affected_formulas:
            return []

        in_degree = {f_id: 0 for f_id in affected_formulas}
        for f_id in affected_formulas:
            expr = self.formulas[f_id]
            for sym in expr.free_symbols:
                sym_name = str(sym)
                for dep_fid in self.graph.get(sym_name, set()):
                    if dep_fid in affected_formulas:
                        in_degree[dep_fid] += 1

        queue = deque(f_id for f_id, deg in in_degree.items() if deg == 0)
        sorted_formulas = []

        while queue:
            fid = queue.popleft()
            sorted_formulas.append(fid)
            for sym in self.formulas[fid].free_symbols:
                sym_name = str(sym)
                for dep_fid in self.graph.get(sym_name, set()):
                    if dep_fid in in_degree:
                        in_degree[dep_fid] -= 1
                        if in_degree[dep_fid] == 0:
                            queue.append(dep_fid)

        for fid in sorted_formulas:
            self.evaluate_formula(fid)

        print('substitution of variables finished with formulas ids: ', sorted_formulas)

        return self.get_triggered_formulas(sorted_formulas)

    def update_variables_topological_Taryan(self, updates) -> None:
        """
        ЕСЛИ будут циклы(более трудные формулы) в графе, тогда простая топологическая сортировка уже не спасет и нужно
        будет использовать Алгоритм Тарьяна(с простыми выражениями он бесполезен, но для циклических даст прирост в скорости
        и будет иметь сложность O(n), как и Кана(проверить потом что быстрее работает) )
        """

    def evaluate_subexpression(self, subexpr): # возвращает float или bool, зависит от sympify().subs()
        """Вычисляет значение подвыражения с учетом кэша"""
        expr = sympify(subexpr)
        vars_values = {str(var): self.variables.get(str(var), 0) for var in expr.free_symbols}
        cache_key = self.get_canonical_cache_key(vars_values)

        if cache_key in self.cache:
            return self.cache[cache_key]

        result = expr.subs(vars_values)
        self.cache[cache_key] = result
        return result

    def evaluate_formula(self, formula_id) -> None:
        """подставляет значение формулы с учетом кэша"""
        expr = self.formulas[formula_id]
        subexpr_values = {}

        # вычисление подвыраженией
        for subexpr in preorder_traversal(expr):
            if subexpr.is_Atom:
                continue
            subexpr_values[str(subexpr)] = self.evaluate_subexpression(str(subexpr))

        # подставление
        expr.subs(subexpr_values).evalf()

    def get_formula_result_by_id(self, formula_id) -> int:  # позже можно сделать чтобы она была частью вебхука для страницы конкретной формулы
        """Возвращает результат вычисления формулы"""

    def evaluate_variable_impact(self, var_name) -> dict: # отслеживать влияние подвыражения смысла пока не вижу, но лишним не будет
        """Оценивает влияние переменной на все выражения и подвыражения."""
        impact = defaultdict(int)
        var_symbol = Symbol(var_name)

        # подвыражения
        for subexpr, count in self.subexpression_weights.items():
            expr_id = self.formula_ids.get(subexpr, None)
            if expr_id is not None:
                expr_sympy = self.formulas[expr_id]
                if var_symbol in expr_sympy.free_symbols:
                    impact[subexpr] += count

        # формулы
        affected_formulas = set()
        for subexpr, formula_ids in self.graph.items():
            if any(var_symbol in self.formulas[f].free_symbols for f in formula_ids):
                affected_formulas.update(formula_ids)

        # обновка
        for formula_id in affected_formulas:
            expr_sympy = self.formulas[formula_id]
            if var_symbol in expr_sympy.free_symbols:
                impact[str(expr_sympy)] += 1

        return dict(impact)

    def is_formula_triggered(self, formula_id) -> bool:
        """возвращает булево значение выражения"""
        expr = self.formulas[formula_id]
        func = self.compiled[formula_id]

        args = {str(var): self.variables.get(str(var), 0) for var in expr.free_symbols}
        cache_key = self.get_canonical_cache_key(args)

        if cache_key in self.cache:
            return self.cache[cache_key]
        else:
            result = func(**args)
            self.cache[cache_key] = result
            return result

    def get_triggered_formulas(self, formula_ids: list[int]) -> list:
        """проверяет, какие из формул сработали после обновления переменных"""
        triggered = []
        for formula_id in formula_ids:
            if self.is_formula_triggered(formula_id):
                triggered.append(formula_id)

        print('Updating finished')
        return triggered

    def get_formulas_variables(self, formula_ids: list) -> (dict[int, list[str]], dict):
        """возвращает подвязанные переменные для каждой формулы из списка ID"""
        result = {}
        variable_values = {}
        for formula_id in formula_ids:
            expr = self.formulas.get(formula_id)
            variables = [str(var) for var in expr.free_symbols]
            result[formula_id] = variables

            for var in variables:
                if var not in variable_values:
                    value = self.variables.get(var)
                    if value is not None:
                        variable_values[var] = value

        return result, variable_values

    @staticmethod
    def get_canonical_cache_key(args):
        """Создаёт уникальный кэш-ключ, не зависящий от порядка переменных"""
        sorted_args = tuple(sorted(args.items()))
        return hashlib.md5(str(sorted_args).encode('utf-8')).hexdigest()

    @staticmethod
    def normalize_expression(expr):
        """Приводит выражение к каноническому виду"""
        if isinstance(expr, Add) or isinstance(expr, Mul):
            return expr.func(*sorted(expr.args, key=str))
        return expr

dependency_graph = DependencyGraph()
