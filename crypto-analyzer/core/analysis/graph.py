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
    0) сделать чтобы float хранился в адекватном виде с двумя знаками после запятой
    1) попробовать сделать чтобы ID формулы было равно ID из базы данных
    2) улучшить добавление(можно позже, чтобы учитывало числа, -> тригонометрию -> скобки -> циклические выражения(самое трудное) ->)
    3) оптимизировать добавление (можно позже)
    """
    def __init__(self):
        self.graph = defaultdict(set)  # {переменная -> набор формул, которые от неё зависят}
        self.formulas = {}  # {ID формулы -> символьное выражение}
        self.compiled = {}  # {ID формулы -> скомпилированная функция}
        self.values = {}  # {переменная -> текущее значение}
        self.cache = {}  # Кэш для промежуточных результатов
        self.formula_ids = {}  # Formula: ID | нужно чтобы нормально удалять их
        self.subexpression_weights = defaultdict(int) # {подвыражение -> количество повторений}

    def add_formula(self, formula_str) -> None: # слишком медленная
        """Добавляет формулу в граф зависимостей"""
        try:
            expr = sympify(formula_str, evaluate=False)
            if not hasattr(expr, "free_symbols"):
                raise ValueError(f"Incorrect expression: {formula_str}")

            formula_id = len(self.formulas)
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
                    f"limit or reduce the permissible severity of the formulas:  {formula_str} \nError\n {e}",
                filename="Graph.log"
            )
            print('опасная формула')
        except Exception as e:
            print(f"Ошибка в формуле '{formula_str}': {e}")

    def remove_formula(self, formula_str) -> None:
        """Удаляет формулу из графа и все связанные подвыражения"""
        try:
            expr = sympify(formula_str, evaluate=False)
            formula_id = self.formula_ids.get(str(expr))

            if formula_id is None:
                custom_logger.log_with_path(
                    level=2,
                    msg=f"The formula was not found, most likely due to a bug in the code, or asynchronous approach:  {formula_str}",
                    filename="Graph.log"
                )
                print(f"Формула '{formula_str}' не найдена.")
                return

            del self.formulas[formula_id]
            del self.compiled[formula_id]
            del self.formula_ids[str(expr)]

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

            print(f"Удалена формула: {formula_str}")
        except Exception as e:
            print(f"Ошибка при удалении формулы '{formula_str}': {e}")

    def update_variables_topological_Kahn(self, updates) -> None:
        """
        алгоритм Кана (не учитывает циклические зависимости)
        Обновляет сразу несколько переменных и пересчитывает только необходимые формулы

        Скорость выполнения зависит от количества уникальных параметров:
        Для 3000 уникальных параметров и 10000 разных формул с количеством пременных равным 3000 обновление
        3000 параметров(по сути всего графа) происхоидит примерно за 0.0004 секунды
        """
        self.values.update(updates)

        affected_formulas = set()
        for var in updates:
            affected_formulas.update(self.graph.get(var, set()))

        # Топологическая сортировка(нереально быстрая)
        in_degree = {f_id: 0 for f_id in affected_formulas}
        for f_id in affected_formulas:
            expr = self.formulas[f_id]
            for var in expr.free_symbols:
                if var in updates:
                    in_degree[f_id] += 1

        queue = deque([f_id for f_id in in_degree if in_degree[f_id] == 0])
        sorted_formulas = []
        while queue:
            f_id = queue.popleft()
            sorted_formulas.append(f_id)
            for var in self.formulas[f_id].free_symbols:
                for dependent_f_id in self.graph.get(str(var), set()):
                    if dependent_f_id in in_degree:
                        in_degree[dependent_f_id] -= 1
                        if in_degree[dependent_f_id] == 0:
                            queue.append(dependent_f_id)

        # пересчет только нужных значений
        for formula_id in sorted_formulas:
            self.evaluate_formula(formula_id)

    def update_variables_topological_Taryan(self, updates) -> None:
        """
        ЕСЛИ будут циклы(более трудные формулы) в графе, тогда топологическая сортировка уже не спасет и нужно
        будет использовать Алгоритм Тарьяна(с простыми выражениями он бесполезен, но для циклических даст прирост в скорости
        и будет иметь сложность O(n), как и топологический(проверить потом что быстрее работает) )
        """

    def evaluate_subexpression(self, subexpr): # возвращает float или bool, зависит от sympify().subs()
        """Вычисляет значение подвыражения с учетом кэша"""
        expr = sympify(subexpr)
        vars_values = {str(var): self.values.get(str(var), 0) for var in expr.free_symbols}
        cache_key = self.get_canonical_cache_key(vars_values)

        if cache_key in self.cache:
            return self.cache[cache_key]

        result = expr.subs(vars_values)
        self.cache[cache_key] = result
        return result

    def evaluate_formula(self, formula_id):
        """Вычисляет значение формулы с учетом кэша"""
        expr = self.formulas[formula_id]
        subexpr_values = {}

        # вычисление подвыраженией
        for subexpr in preorder_traversal(expr):
            if subexpr.is_Atom:
                continue
            subexpr_values[str(subexpr)] = self.evaluate_subexpression(str(subexpr))

        # подставление
        return expr.subs(subexpr_values).evalf()

    def get_formula_result_by_id(self, formula_id) -> int:  # возможно не нужна
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
        '''
        позже может быть использована для пересчета "забущенных" выражений.
         Нужно сделать функцию для перерасчета сразу всех значений или улучшить обновление(более вероятно)
        '''
        expr = self.formulas[formula_id]
        func = self.compiled[formula_id]

        args = {str(var): self.values.get(str(var), 0) for var in expr.free_symbols}
        cache_key = self.get_canonical_cache_key(args)

        if cache_key in self.cache:
            return self.cache[cache_key]
        else:
            result = func(**args)
            self.cache[cache_key] = result
            return result

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
