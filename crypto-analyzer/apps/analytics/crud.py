from datetime import datetime

from sqlalchemy import select, func
from sqlalchemy.ext.asyncio import AsyncSession
from sqlalchemy.orm import selectinload

from core.models import TriggerFormula, CryptoApi, CryptoParams, TriggerFormulaComponent, TriggerComponent, \
    CryptoCurrencies, TriggerHistory, TriggerComponentsHistory
from db.postgre import postgres_db


async def get_formula_by_id(session: AsyncSession, pk: int, *fields) -> str:
    query = await session.execute(select(*fields).where(
        TriggerFormula.id == pk,
    ))
    result = query.scalar()

    return result


async def get_actual_formulas():
    async with postgres_db.session_factory() as session:
        query = await session.execute(select(TriggerFormula.id, TriggerFormula.formula).where(
            TriggerFormula.is_active == True,
            TriggerFormula.is_shutted_off == False,
        ))
        result = query.fetchall()

    return result

async def get_actual_components() -> dict: # возможная оптимизация - добавть formulas_num в CryptoApi и относительно его уже плясать
    """получает необходимые апи, к которым нужно делать запрос в зависимости от актальности формул и компонентов"""
    async with postgres_db.session_factory() as session:
        query = await session.execute(
            select(CryptoApi.api, func.count().label("count"))
            .join(TriggerComponent, CryptoApi.id == TriggerComponent.api_id)
            .join(CryptoParams, TriggerComponent.parameter_id == CryptoParams.id)
            .join(TriggerFormulaComponent, TriggerFormulaComponent.component_id == TriggerComponent.id)
            .join(TriggerFormula, TriggerFormula.id == TriggerFormulaComponent.formula_id)
            .where(
                CryptoApi.is_actual.is_(True),
                CryptoParams.is_active.is_(True),
                TriggerFormula.is_active.is_(True),
            )
            .group_by(CryptoApi.api)
        )
        rows = query.all()

        return {api: count for api, count in rows}

async def get_needed_fields_from_endpoint(endpoint: str) -> dict[str, list[str]]:
    """получает список необходимых параметров для конкретного эндпоинта"""
    async with postgres_db.session_factory() as session:
        query = (
            select(CryptoCurrencies.currency, CryptoParams.parameter)
            .join(TriggerComponent, TriggerComponent.parameter_id == CryptoParams.id)
            .join(CryptoCurrencies, TriggerComponent.currency_id == CryptoCurrencies.id)
            .join(CryptoApi, CryptoApi.id == TriggerComponent.api_id)
            .join(TriggerFormulaComponent, TriggerFormulaComponent.component_id == TriggerComponent.id)
            .join(TriggerFormula, TriggerFormula.id == TriggerFormulaComponent.formula_id)
            .where(
                CryptoApi.api == endpoint,
                CryptoApi.is_actual.is_(True),
                TriggerFormula.is_active.is_(True),
                TriggerFormula.is_shutted_off.is_(False),
                CryptoParams.is_active.is_(True)
            )
        )

        result = await session.execute(query)
        rows = result.all()

    needed_fields = {}
    for symbol, parameter in rows:
        needed_fields.setdefault(symbol, []).append(parameter)

    return needed_fields


#  позже сделать возможность более детального учета - чтобы данная функция срабатывала всегда, это добавит много нагрузки
async def add_trigger_history(data: dict[int, list[str]], formulas_values: dict[str, float]):
    """записывает в историю сопутствующие данные сработавших триггеров"""
    print(data, formulas_values)
    async with postgres_db.session_factory() as session:
        async with session.begin():
            timestamp = datetime.utcnow()

            for formula_id, variable_names in data.items():
                trigger_history = TriggerHistory(
                    formula_id=formula_id,
                    timestamp=timestamp,
                    status=True  # пока что всегда правда
                )

                session.add(trigger_history)
                await session.flush()

                result = await session.execute(
                    select(TriggerFormulaComponent)
                    .options(selectinload(TriggerFormulaComponent.component))
                    .where(TriggerFormulaComponent.formula_id == formula_id)
                )
                components = result.scalars().all()

                comp_map: dict[str, TriggerFormulaComponent] = {
                    comp.component.name: comp
                    for comp in components
                }
                for var_name in variable_names:
                    value = formulas_values.get(var_name)
                    if value is None:
                        continue

                    component = comp_map.get(var_name)
                    if not component:
                        continue

                    component_history = TriggerComponentsHistory(
                        trigger_history_id=trigger_history.id,
                        component_id=component.id,
                        value=value
                    )

                    session.add(component_history)
            await session.commit()
