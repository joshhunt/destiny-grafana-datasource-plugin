import { uniqBy } from 'lodash';

import React, { useCallback, useEffect, useMemo, useState } from 'react';
import { AsyncSelect, Select } from '@grafana/ui';
import { QueryEditorProps, SelectableValue } from '@grafana/data';
import { DataSource } from '../datasource';
import {
  CharacterItem as ListCharactersItem,
  Membership,
  MyDataSourceOptions,
  MyQuery,
  TrialsReportSearchResult,
} from '../types';
import { EditorField, EditorRow, EditorRows, EditorSwitch } from '@grafana/experimental';

type Props = QueryEditorProps<DataSource, MyQuery, MyDataSourceOptions>;

export function QueryEditor({ query, onChange, onRunQuery, datasource }: Props) {
  const [characterOptions, setCharacterOptions] = useState<ListCharactersItem[]>([]);
  const [activityModes, setActivityModes] = useState<SelectableValue[]>([]);
  const [isSearching, setIsSearching] = useState(false);

  const updateQuery = useCallback(
    (update: Partial<MyQuery>) => {
      const newQuery = { ...query, ...update };
      onChange(newQuery);

      if (newQuery.profile) {
        onRunQuery();
      }
    },
    [query, onChange, onRunQuery]
  );

  /**
   * Request character options on profile change
   */
  useEffect(() => {
    if (!query.profile) {
      return;
    }

    datasource.postResource<ListCharactersItem[]>('list-characters', query.profile).then((characters) => {
      setCharacterOptions(characters);
    });
  }, [datasource, query.profile]);

  /**
   * Request activity modes on load
   */
  useEffect(() => {
    datasource.getResource<SelectableValue[]>('list-activity-modes').then((characters) => {
      setActivityModes(characters);
    });
  }, [datasource, query.profile]);

  /**
   * Ensure character list contains only valid characters
   */
  useEffect(() => {
    if (!characterOptions.length) {
      return;
    }
    const selectedCharacters = query.characters ?? [];

    for (const characterId of selectedCharacters) {
      const isValidCharacter = characterOptions.some((v) => v.characterId === characterId);
      if (!isValidCharacter) {
        updateQuery({ characters: [] });
        break;
      }
    }
  }, [characterOptions, query.characters, updateQuery]);

  const loadProfileSearchOptions = useCallback(
    async (query: string): Promise<Array<SelectableValue<Membership>>> => {
      let results = await datasource.postResource<TrialsReportSearchResult[]>('profile-search', { query });
      results = uniqBy(results, (v) => v.bungieName);

      return results.map((v) => {
        let membershipId = v.membershipId;
        let membershipType = v.membershipType;

        if (v.crossSaveOverride) {
          membershipId = v.crossSaveOverride.membershipId;
          membershipType = v.crossSaveOverride.membershipType;
        }

        return {
          label: v.bungieName,
          value: { membershipId, membershipType, bungieName: v.bungieName },
        };
      });
    },
    [datasource]
  );

  const handleSearchInputChange = useCallback((searchValue: string) => setIsSearching(!!searchValue), []);

  const onMembershipChange = useCallback(
    (change: SelectableValue<Membership>) => {
      updateQuery({ profile: change.value });
    },
    [updateQuery]
  );

  const onActivityModeChange = useCallback(
    (change: SelectableValue | undefined) => {
      updateQuery({ activityMode: change?.value });
    },
    [updateQuery]
  );

  const onCharacterToggled = useCallback(
    (characterId: string, newValue: boolean) => {
      let characters = query.characters ?? [];

      if (newValue) {
        characters = [...characters, characterId];
      } else {
        characters = characters.filter((v) => v !== characterId);
      }

      characters = characters.filter((v, index, arr) => {
        const isValidCharacterId = characterOptions.some((opt) => opt.characterId === v);
        return arr.indexOf(v) === index && isValidCharacterId;
      });

      updateQuery({ characters });
    },
    [characterOptions, query.characters, updateQuery]
  );

  const profileValue = useMemo(() => {
    if (!query.profile) {
      return [];
    }

    return [
      {
        label: query.profile.bungieName,
        value: query.profile,
      },
    ];
  }, [query.profile]);

  const charactersToRender = useMemo(() => {
    return characterOptions.length
      ? characterOptions
      : [
          { description: 'Warlock', characterId: '1', isPlaceholder: true },
          { description: 'Hunter', characterId: '2', isPlaceholder: true },
          { description: 'Titan', characterId: '3', isPlaceholder: true },
        ];
  }, [characterOptions]);

  return (
    <EditorRows>
      <EditorRow>
        <EditorField label="Player">
          <AsyncSelect
            width={26}
            loadOptions={loadProfileSearchOptions}
            onChange={onMembershipChange}
            value={profileValue[0]}
            defaultOptions={profileValue}
            onInputChange={handleSearchInputChange}
            noOptionsMessage={isSearching ? 'No players found' : 'Type to search for player'}
            loadingMessage="Searching..."
          />
        </EditorField>

        {charactersToRender.map((v) => {
          return (
            <EditorField key={v.characterId} width={8} label={v.description}>
              <>
                <EditorSwitch
                  disabled={v.isPlaceholder}
                  value={query.characters?.includes(v.characterId)}
                  onChange={(ev) => !v.isPlaceholder && onCharacterToggled(v.characterId, ev.currentTarget.checked)}
                />
              </>
            </EditorField>
          );
        })}

        <EditorField label="Activity mode">
          <Select
            value={query.activityMode}
            width={30}
            options={activityModes}
            onChange={onActivityModeChange}
            isClearable
          />
        </EditorField>
      </EditorRow>
    </EditorRows>
  );
}
