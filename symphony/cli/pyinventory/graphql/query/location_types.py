#!/usr/bin/env python3
# @generated AUTOGENERATED file. Do not Change!

from dataclasses import dataclass
from datetime import datetime
from gql.gql.datetime_utils import DATETIME_FIELD
from gql.gql.graphql_client import GraphqlClient
from gql.gql.client import OperationException
from gql.gql.reporter import FailedOperationException
from functools import partial
from numbers import Number
from typing import Any, Callable, List, Mapping, Optional
from time import perf_counter
from dataclasses_json import DataClassJsonMixin

from ..fragment.property_type import PropertyTypeFragment, QUERY as PropertyTypeFragmentQuery

QUERY: List[str] = PropertyTypeFragmentQuery + ["""
query LocationTypesQuery {
  locationTypes {
    edges {
      node {
        id
        name
        propertyTypes {
          ...PropertyTypeFragment
        }
      }
    }
  }
}

"""]

@dataclass
class LocationTypesQuery(DataClassJsonMixin):
    @dataclass
    class LocationTypesQueryData(DataClassJsonMixin):
        @dataclass
        class LocationTypeConnection(DataClassJsonMixin):
            @dataclass
            class LocationTypeEdge(DataClassJsonMixin):
                @dataclass
                class LocationType(DataClassJsonMixin):
                    @dataclass
                    class PropertyType(PropertyTypeFragment):
                        pass

                    id: str
                    name: str
                    propertyTypes: List[PropertyType]

                node: Optional[LocationType]

            edges: List[LocationTypeEdge]

        locationTypes: Optional[LocationTypeConnection]

    data: LocationTypesQueryData

    @classmethod
    # fmt: off
    def execute(cls, client: GraphqlClient) -> Optional[LocationTypesQueryData.LocationTypeConnection]:
        # fmt: off
        variables = {}
        try:
            start_time = perf_counter()
            response_text = client.call(''.join(set(QUERY)), variables=variables)
            res = cls.from_json(response_text).data
            elapsed_time = perf_counter() - start_time
            client.reporter.log_successful_operation("LocationTypesQuery", variables, elapsed_time)
            return res.locationTypes
        except OperationException as e:
            raise FailedOperationException(
                client.reporter,
                e.err_msg,
                e.err_id,
                "LocationTypesQuery",
                variables,
            )