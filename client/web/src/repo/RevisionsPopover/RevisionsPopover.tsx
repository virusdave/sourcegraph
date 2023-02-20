import React, { useCallback, useEffect, useMemo } from 'react'

import { mdiClose } from '@mdi/js'

import { GitRefType, Scalars } from '@sourcegraph/shared/src/graphql-operations'
import { Button, useLocalStorage, Tab, TabList, TabPanel, TabPanels, Icon } from '@sourcegraph/wildcard'

import { GitCommitAncestorFields, GitRefFields } from '../../graphql-operations'
import { eventLogger } from '../../tracking/eventLogger'
import { replaceRevisionInURL } from '../../util/url'

import { ConnectionPopoverTabs } from './components'
import { RevisionsPopoverCommits } from './RevisionsPopoverCommits'
import { RevisionsPopoverReferences } from './RevisionsPopoverReferences'

import styles from './RevisionsPopover.module.scss'
import { RepoType } from '../RepoRevisionContainer'

export interface RevisionsPopoverProps {
    repoId: Scalars['ID']
    repoName: string
    defaultBranch: string

    /** The current revision, or undefined for the default branch. */
    currentRev: string | undefined

    currentCommitID?: string

    /* Callback to dismiss the parent popover wrapper */
    togglePopover: () => void

    /* Determine the URL to use for each revision node */
    getPathFromRevision?: (href: string, revision: string) => string

    /**
     * If the popover should display result nodes that are not **known** revisions.
     * This ensures we can support ancestory-based revision queries (e.g. `main^1`).
     */
    showSpeculativeResults?: boolean

    /**
     * The selected revision node. Should be used to trigger side effects from clicking a node, e.g. calling `eventLogger`.
     */
    onSelect?: (node: GitRefFields | GitCommitAncestorFields) => void

    repoType: RepoType
}

type RevisionsPopoverTabID = 'branches' | 'tags' | 'commits'

interface RevisionsPopoverTab {
    id: RevisionsPopoverTabID
    label: string
    noun: string
    pluralNoun: string
    type?: GitRefType
    description: string
}

const LAST_TAB_STORAGE_KEY = 'RevisionsPopover.lastTab'

const BRANCHES_TAB: RevisionsPopoverTab = {
    id: 'branches',
    label: 'Branches',
    noun: 'branch',
    pluralNoun: 'branches',
    type: GitRefType.GIT_BRANCH,
    description: 'Find a revision from the listed branches',
}
const TAGS_TAB: RevisionsPopoverTab = {
    id: 'tags',
    label: 'Tags',
    noun: 'tag',
    pluralNoun: 'tags',
    type: GitRefType.GIT_TAG,
    description: 'Find a revision from the listed tags',
}
const COMMITS_TAB: RevisionsPopoverTab = {
    id: 'commits',
    label: 'Commits',
    noun: 'commit',
    pluralNoun: 'commits',
    description: 'Find a revision from the listed commits',
}

/**
 * A popover that displays a searchable list of revisions (grouped by type) for
 * the current repository.
 */
export const RevisionsPopover: React.FunctionComponent<React.PropsWithChildren<RevisionsPopoverProps>> = props => {
    const { getPathFromRevision = replaceRevisionInURL, repoType } = props

    useEffect(() => {
        eventLogger.logViewEvent('RevisionsPopover')
    }, [])

    const [tabIndex, setTabIndex] = useLocalStorage(LAST_TAB_STORAGE_KEY, 0)
    const handleTabsChange = useCallback((index: number) => setTabIndex(index), [setTabIndex])

    const tabs = useMemo(() => {
        if (repoType === 'packageRepo') {
            return [TAGS_TAB]
        }

        return [BRANCHES_TAB, TAGS_TAB, COMMITS_TAB]
    }, [repoType])

    return (
        <ConnectionPopoverTabs
            className={styles.revisionsPopover}
            data-testid="revisions-popover"
            defaultIndex={tabs.length === 1 ? 0 : tabIndex}
            onChange={handleTabsChange}
        >
            <TabList
                wrapperClassName={styles.tabs}
                actions={
                    <Button
                        onClick={props.togglePopover}
                        variant="icon"
                        className={styles.tabsClose}
                        aria-label="Close"
                    >
                        <Icon aria-hidden={true} svgPath={mdiClose} />
                    </Button>
                }
            >
                {tabs.map(({ label, id }) => (
                    <Tab key={id} data-tab-content={id}>
                        <span className="tablist-wrapper--tab-label">{label}</span>
                    </Tab>
                ))}
            </TabList>
            <TabPanels>
                {tabs.map(tab => (
                    <TabPanel key={tab.id} tabIndex={-1}>
                        {tab.type ? (
                            <RevisionsPopoverReferences
                                noun={tab.noun}
                                pluralNoun={tab.pluralNoun}
                                type={tab.type}
                                currentRev={props.currentRev}
                                getPathFromRevision={getPathFromRevision}
                                defaultBranch={props.defaultBranch}
                                repo={props.repoId}
                                repoName={props.repoName}
                                onSelect={props.onSelect}
                                showSpeculativeResults={
                                    props.showSpeculativeResults && tab.type === GitRefType.GIT_BRANCH
                                }
                                tabLabel={tab.description}
                            />
                        ) : (
                            <RevisionsPopoverCommits
                                noun={tab.noun}
                                pluralNoun={tab.pluralNoun}
                                currentRev={props.currentRev}
                                getPathFromRevision={getPathFromRevision}
                                defaultBranch={props.defaultBranch}
                                repo={props.repoId}
                                currentCommitID={props.currentCommitID}
                                onSelect={props.onSelect}
                                tabLabel={tab.description}
                            />
                        )}
                    </TabPanel>
                ))}
            </TabPanels>
        </ConnectionPopoverTabs>
    )
}
