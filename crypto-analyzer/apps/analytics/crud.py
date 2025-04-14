from sqlalchemy import select, func
from sqlalchemy.ext.asyncio import AsyncSession

from core.models import TriggerFormula, CryptoApi, CryptoParams, TriggerFormulaComponent, TriggerComponent, \
    CryptoCurrencies
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
