import React from "react";
import { FormControl, InputLabel, MenuItem, Select } from "@mui/material";
import { DrawingRepoRef } from "./drawingAPI";

import style from "./RepositorySelector.module.css";

interface RepositorySelectorProps {
	readonly availableRepos: DrawingRepoRef[];
	readonly currentSelection: DrawingRepoRef;
	readonly disabled?: boolean;
	readonly requestSelectionChange: (requestedRepo: DrawingRepoRef) => void;
}

export const RepositorySelector = ({ availableRepos, currentSelection, disabled, requestSelectionChange }: RepositorySelectorProps) => {
	return <FormControl className={style.repositorySelector} disabled={disabled ?? false}>
		<InputLabel>Repository</InputLabel>
		<Select
			label="Repository"
			onChange={event => {
				const repo = availableRepos.find(repo => repo.name === event.target.value) ?? currentSelection;
				requestSelectionChange(repo);
			}}
			value={currentSelection?.name || null}
		>
			{
				availableRepos.map(repo => <MenuItem key={repo.name} value={repo.name}>{repo.label}</MenuItem>)
			}
		</Select>
	</FormControl>;
};
