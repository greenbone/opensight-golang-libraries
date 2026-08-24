# Full-text search specification

## Purpose

This document defines the behavior of the `TextContains` compare operator for
English-language full-text search. It specifies the public input contract, the
required OpenSearch field mapping, the generated query semantics, and the
observable matching guarantees.

The key words **MUST**, **MUST NOT**, **SHOULD**, **SHOULD NOT**, and **MAY** in
this document indicate requirement levels.

## Scope

`TextContains` combines two search mechanisms:

- English full-text matching on a primary field, including stemming; and
- exact in-word substring matching on an `<field>.ngram` multi-field.

For a multi-unit input, each meaningful unit MUST match, but each unit MAY use
either mechanism independently. For example, `PLC ttings spotted` can match
`PLC settings or the upload of malicious programs spot`: `PLC` matches the
primary field, `ttings` matches the n-gram field, and `spotted` matches `spot`
through English stemming.

Fuzzy matching is deliberately excluded. Combined with linguistic analysis
such as stemming, it broadens the result set with too many irrelevant matches.
`TextContains` prioritizes precise full-text and substring matches over typo
tolerance.

This specification does not require:

- fuzzy or typo-tolerant matching;
- matches for substrings shorter than 3 or longer than 10 characters;
- synonym or semantic-equivalence matching;
- a particular relevance order between primary-field and n-gram matches; or
- creation or migration of service-owned OpenSearch mappings by this library.

## Input contract

The operator accepts either:

- one non-empty, non-whitespace string; or
- a non-empty array in which every element is a non-empty, non-whitespace
  string.

All other values MUST be rejected. This includes non-string scalars, empty
strings, whitespace-only strings, empty arrays, and arrays containing an
invalid element. Array elements MUST be validated in order, and processing
MUST stop at the first invalid element.

For query construction, each string value MUST be split into units at runs of
Unicode whitespace, equivalent to Go's `strings.Fields`. OpenSearch, rather
than the Go query builder, MUST continue to perform lowercasing, stemming,
punctuation handling, ASCII folding, and stopword removal.

## Index contract

### Field structure

A field used with `TextContains` MUST be mapped as `text`. To provide all
behavior in this specification, it MUST also have a text multi-field named
`ngram`:

```text
<field>        primary English full-text field
<field>.ngram  exact substring field
```

The primary field MUST NOT be replaced by an n-gram field. The primary field
provides English morphology; the n-gram multi-field provides
in-word substring matching.

If `<field>.ngram` is absent, the query MUST remain usable through the primary
field, but substring behavior is unavailable.

### Analyzer requirements

The primary analyzer MUST:

- tokenize with the standard tokenizer;
- lowercase tokens;
- handle English possessives and English stemming; and
- remove the configured English stopwords while preserving `no` and `not`.

The n-gram index analyzer MUST:

- tokenize with the standard tokenizer;
- lowercase and ASCII-fold tokens;
- apply the same negation-preserving stopword policy as the primary analyzer;
- create grams from 3 through 10 characters; and
- preserve the original normalized token.

The n-gram search analyzer MUST apply the same tokenization, normalization, and
stopword policy without creating n-grams.

The index setting `index.max_ngram_diff` MUST be at least `7`, which is the
difference between the required maximum and minimum gram lengths.

The following mapping is conforming. `content` is illustrative and MUST be
replaced with the actual field name:

```json
{
  "settings": {
    "index.max_ngram_diff": 7,
    "analysis": {
      "filter": {
        "english_stop_without_negation": {
          "type": "stop",
          "stopwords": [
            "a", "an", "and", "are", "as", "at", "be", "but", "by",
            "for", "if", "in", "into", "is", "it", "of", "on", "or",
            "such", "that", "the", "their", "then", "there", "these",
            "they", "this", "to", "was", "will", "with"
          ]
        },
        "english_possessive_stemmer": {
          "type": "stemmer",
          "language": "possessive_english"
        },
        "english_stemmer": {
          "type": "stemmer",
          "language": "english"
        },
        "substring_grams": {
          "type": "ngram",
          "min_gram": 3,
          "max_gram": 10,
          "preserve_original": true
        }
      },
      "analyzer": {
        "english_preserving_negation": {
          "type": "custom",
          "tokenizer": "standard",
          "filter": [
            "english_possessive_stemmer",
            "lowercase",
            "english_stop_without_negation",
            "english_stemmer"
          ]
        },
        "substring_index": {
          "type": "custom",
          "tokenizer": "standard",
          "filter": [
            "lowercase",
            "asciifolding",
            "english_stop_without_negation",
            "substring_grams"
          ]
        },
        "substring_search": {
          "type": "custom",
          "tokenizer": "standard",
          "filter": [
            "lowercase",
            "asciifolding",
            "english_stop_without_negation"
          ]
        }
      }
    }
  },
  "mappings": {
    "properties": {
      "content": {
        "type": "text",
        "analyzer": "english_preserving_negation",
        "fields": {
          "ngram": {
            "type": "text",
            "analyzer": "substring_index",
            "search_analyzer": "substring_search"
          }
        }
      }
    }
  }
}
```

Equivalent analyzer and filter names MAY be used, but their observable behavior
MUST satisfy these requirements.

### Substring boundaries

Correct in-word substrings from 3 through 10 characters MUST be searchable on
the n-gram field. One- and two-character inputs MUST NOT gain substring
behavior from that field. An input longer than 10 characters MUST NOT match a
longer token merely because it occurs inside that token.

Complete tokens longer than 10 characters remain searchable through the
primary field and through the original token retained by the n-gram analyzer.

## Query contract

### Single value

For one valid string value `v`, let `units(v)` be its whitespace-delimited
units. The generated query MUST be semantically equivalent to:

```text
guard(v)
AND for every unit u in units(v): unit_match(u)

guard(v):
  primary(v, operator=OR, zero_terms_query=NONE)
  OR exact_ngram(v, operator=OR, zero_terms_query=NONE)

unit_match(u):
  primary(u, operator=AND, zero_terms_query=ALL)
  OR exact_ngram(u, operator=AND, zero_terms_query=ALL)
```

Every `OR` group MUST require at least one of its two alternatives. The guard
and every unit group MUST be required by an outer Boolean `must` query.

The guard ensures that the complete input contains at least one meaningful
analyzed term. The per-unit groups then require every meaningful unit while
allowing different units to select different matching mechanisms.

### Match-clause parameters

Primary-field clauses MUST use these parameters:

| Parameter | Guard | Per-unit clause |
| --- | --- | --- |
| `operator` | `or` | `and` |
| `zero_terms_query` | `none` | `all` |

N-gram clauses MUST use these parameters:

| Parameter | Guard | Per-unit clause |
| --- | --- | --- |
| `operator` | `or` | `and` |
| `zero_terms_query` | `none` | `all` |
| `fuzziness` | disabled | disabled |

`zero_terms_query: none` MAY be omitted from the serialized guard because it is
the OpenSearch default. Fuzziness MUST NOT be enabled on either the primary
field or `<field>.ngram`; neither misspelled words nor misspelled substrings are
supported.

### Stopwords and negation

`zero_terms_query: all` makes a removed stopword neutral in an individual unit
group. The guard's `zero_terms_query: none` ensures that an input containing
only removed stopwords matches no documents.

The index and search analyzers MUST preserve `no` and `not`. These words are
meaningful units and MUST NOT be treated as neutral stopwords.

### Multiple values

For an array value, the operator MUST construct one complete single-value query
for each element. Those queries MUST be combined with Boolean `should`, with
`minimum_should_match: 1`. Array elements therefore have OR semantics, while
the meaningful units within each element retain AND semantics.

## Required observable behavior

A conforming implementation and mapping MUST satisfy the following cases:

| Indexed text | Search input | Expected result | Reason |
| --- | --- | --- | --- |
| `Memory leaks` | `memory leak` | match | English stemming |
| `Memory leaks` | `memories leaking` | match | English inflection handling |
| `Memory leaks` | `memroy leak` | no match | Typo-tolerant matching is disabled |
| `the Microarchitecture` | `roarch` | match | In-word substring within bounds |
| `the Microarchitecture` | `roa` | match | Minimum n-gram boundary |
| `the Microarchitecture` | `roarchitec` | match | Maximum n-gram boundary |
| `the Microarchitecture` | `ro` | no match | Below minimum substring length |
| `the Microarchitecture` | `roarchitect` | no match | Above maximum substring length |
| `Memory leaks` | `memory unrelated` | no match | Every meaningful unit is required |
| `Memory leaks` | `the memory` | match | Removed stopword is neutral |
| `Memory leaks` | `the` | no match | Guard rejects stopword-only input |
| `This system is not vulnerable` | `not vulnerable` | match | Negation is preserved |
| `the Microarchitecture` | `raorch` | no match | N-gram matching is not fuzzy |
| `PLC settings or the upload of malicious programs spot` | `PLC ttings spotted` | match | Substring and stemming mechanisms are combined per unit |
| `PLC settings or the upload of malicious programs spot` | `PLC ttings spooted` | no match | Misspelled units do not match |
| `PLC settings or the upload of malicious programs spot` | `PLC ttings unrelated` | no match | Mixed queries still require every meaningful unit |

Matching is case-insensitive. Simple punctuation inside a whitespace-delimited
unit MUST be interpreted by the configured OpenSearch analyzers rather than by
special preprocessing in the query builder.

Algorithmic English stemming does not guarantee irregular forms or semantic
equivalents such as `child` and `children`. Synonyms MAY be added by a consuming
service, but they are outside this contract.

## Mapping lifecycle and deployment

The service that owns an OpenSearch index is responsible for satisfying the
index contract. Adding or changing analyzers or adding the `ngram` multi-field
does not populate that field for existing indexed values. The service MUST
create a compatible index and reindex existing documents before it relies on
substring matching.

During a rolling migration, a service MAY deploy the query behavior before all
indexes contain the n-gram field because the primary-field alternative remains
usable. It MUST NOT claim substring conformance until the mapping is deployed
and documents are reindexed.

Before deployment, the consuming service SHOULD:

- verify both custom analyzers with the OpenSearch Analyze API;
- test all required observable behaviors against a newly created index;
- measure representative index-size growth and query latency; and
- bound the number of accepted search units so query expansion cannot exceed
  the cluster's Boolean-clause limit.

## Conformance

A `TextContains` implementation conforms to this specification only if:

- its public validation and array behavior satisfy the input contract;
- its generated query is semantically equivalent to the query contract;
- it does not enable fuzziness; and
- its integration tests cover the required observable behavior.

A consuming service provides fully conforming substring search only if its
field mapping and indexed documents also satisfy the index contract.

## References

- [OpenSearch match query](https://docs.opensearch.org/latest/query-dsl/full-text/match/)
- [OpenSearch English analyzer](https://docs.opensearch.org/latest/analyzers/language-analyzers/english/)
- [OpenSearch n-gram token filter](https://docs.opensearch.org/latest/analyzers/token-filters/ngram/)
- [OpenSearch multi-fields](https://docs.opensearch.org/latest/mappings/mapping-parameters/fields/)
