import { isNil } from "lodash";
import { createAppSlice } from "../../app/createAppSlice";
import { emptyArray } from "../../utils/empty-array";
import type { Drawing, DrawingLists, DrawingRepoRef, FQDrawingId } from "./drawingAPI";
import { fetchDrawing, fetchDrawingList, createDrawing as createDrawingApi, saveDrawing, fetchDrawingRepositories, fqDrawingIdToString } from "./drawingAPI";

export enum AsyncOperationState {
	idle = "idle",
	inProgress = "inProgress",
	failed = "failed"
};

const emptyDrawing: Drawing = {
	id: "",
	title: "",
	elements: emptyArray,
	repo: {
		name: "",
		label: ""
	}
};

export interface DrawingSliceState {
	drawingRepos: {
		getRepos: {
			status: AsyncOperationState;
		},
		value: DrawingRepoRef[]
	},
	drawingLists: {
		getList: {
			status: AsyncOperationState;
		}
		value: DrawingLists;
	};
	drawingInEdit: {
		open: {
			status: AsyncOperationState;
		}
		save: {
			status: AsyncOperationState;
		};
		savedDrawing: Drawing;
	};
	drawingManagement: {
		rename: {
			status: AsyncOperationState
		}
		delete: {
			status: AsyncOperationState
		}
	}
}

const initialState: DrawingSliceState = {
	drawingRepos: {
		getRepos: {
			status: AsyncOperationState.idle
		},
		value: []
	},
	drawingLists: {
		getList: {
			status: AsyncOperationState.idle
		},
		value: {}
	},

	drawingInEdit: {
		open: {
			status: AsyncOperationState.idle
		},
		save: {
			status: AsyncOperationState.idle
		},
		savedDrawing: emptyDrawing
	},
	drawingManagement: {
		rename: {
			status: AsyncOperationState.idle
		},
		delete: {
			status: AsyncOperationState.idle
		}
	}
};

const setLocation = (fQDrawingId: FQDrawingId | null) => {
	if (!isNil(fQDrawingId)) {
		const drawingId = fqDrawingIdToString(fQDrawingId);
		const desiredPathname = `/drawings/${drawingId}`;
		if (window.location.pathname !== desiredPathname) {
			window.history.pushState({}, drawingId, desiredPathname);
		}
	} else {
		if (window.location.pathname !== "") {
			window.history.pushState({}, "", "/");
		}
	}
};

// If you are not using async thunks you can use the standalone `createSlice`.
export const drawingSlice = createAppSlice({
	name: "drawing",
	initialState,
	reducers: create => ({
		getDrawingRepositories: create.asyncThunk(
			async () => {
				return await fetchDrawingRepositories();
			},
			{
				pending: state => {
					state.drawingRepos.getRepos.status = AsyncOperationState.inProgress;
				},
				fulfilled: (state, action) => {
					state.drawingRepos.getRepos.status = AsyncOperationState.idle;
					state.drawingRepos.value = action.payload;
				},
				rejected: state => {
					state.drawingRepos.getRepos.status = AsyncOperationState.failed;
				}
			}

		),
		getDrawingLists: create.asyncThunk(
			async () => {
				return await fetchDrawingList();
			},
			{
				pending: state => {
					state.drawingLists.getList.status = AsyncOperationState.inProgress;
				},
				fulfilled: (state, action) => {
					state.drawingLists.getList.status = AsyncOperationState.idle;
					state.drawingLists.value = action.payload;
				},
				rejected: state => {
					state.drawingLists.getList.status = AsyncOperationState.failed;
				}
			}
		),
		getDrawingContent: create.asyncThunk(
			async (fqDrawingId: FQDrawingId) => {
				const response = await fetchDrawing(fqDrawingId);
				setLocation(fqDrawingId);
				return response;
			},
			{
				pending: state => {
					state.drawingInEdit.open.status = AsyncOperationState.inProgress;
				},
				fulfilled: (state, action) => {
					return {
						...state,
						drawingInEdit: {
							...state.drawingInEdit,
							open: {
								...state.drawingInEdit.open,
								status: AsyncOperationState.idle
							},
							savedDrawing: {
								...action.payload,
								id: action.meta.arg.drawingId,
								repo: {
									name: action.meta.arg.repoId,
									label: state.drawingRepos.value.find(repo => repo.name === action.meta.arg.repoId)?.label ?? "???"
								}
							},
							currentElements: action.payload.elements
						}
					};
				},
				rejected: state => {
					state.drawingInEdit.open.status = AsyncOperationState.failed;
				}
			}
		),
		clearCanvas: create.preparedReducer(() => {
			setLocation(null);
			// if (window.location.pathname !== "") {
			// 	window.history.pushState({}, "", "/");
			// }
			return {
				payload: undefined
			};
		}, state => {
			return {
				...state,
				drawingInEdit: {
					...state.drawingInEdit,
					savedDrawing: emptyDrawing,
					currentElements: emptyArray
				}
			};
		}),
		createDrawing: create.asyncThunk(
			async (payload: Drawing) => {
				const drawingId = await createDrawingApi(payload.title, payload.elements);
				setLocation({ repoId: payload.repo!.name!, drawingId: payload.id! });
				return drawingId;
			},
			{
				pending: state => {
					state.drawingInEdit.save.status = AsyncOperationState.inProgress;
				},
				fulfilled: (state, action) => {
					return {
						...state,
						drawingInEdit: {
							...state.drawingInEdit,
							savedDrawing: {
								...action.meta.arg,
								id: action.payload
							},
							save: {
								...state.drawingInEdit.save,
								status: AsyncOperationState.idle
							}
						}
					};
				},
				rejected: state => {
					state.drawingInEdit.save.status = AsyncOperationState.failed;
				}
			}
		),
		saveDrawingContent: create.asyncThunk(
			async (payload: Drawing) => {
				await saveDrawing(payload);
			},
			{
				pending: state => {
					state.drawingInEdit.save.status = AsyncOperationState.inProgress;
				},
				fulfilled: (state, action) => {
					return {
						...state,
						drawingInEdit: {
							...state.drawingInEdit,
							savedDrawing: action.meta.arg,
							save: {
								...state.drawingInEdit.save,
								status: AsyncOperationState.idle
							}
						}
					};
				},
				rejected: state => {
					state.drawingInEdit.save.status = AsyncOperationState.failed;
				}
			}
		)
	}),
	// You can define your selectors here. These selectors receive the slice
	// state as their first argument.
	selectors: {
		selectDrawingRepos: drawing => drawing.drawingRepos.value,
		selectDrawingLists: drawing => drawing.drawingLists.value,
		selectDrawingListStatus: drawing => drawing.drawingLists.getList.status,
		selectSavedDrawing: drawing => drawing.drawingInEdit.savedDrawing,
		selectDrawingToEditStatus: drawing => drawing.drawingInEdit.open.status,
		selectSaveDrawingStatus: drawing => drawing.drawingInEdit.save.status
	}
});

export const { getDrawingRepositories, getDrawingLists, getDrawingContent, clearCanvas, createDrawing, saveDrawingContent } =
	drawingSlice.actions;

export const {
	selectDrawingRepos,
	selectDrawingLists,
	selectDrawingListStatus,
	selectSavedDrawing,
	selectSaveDrawingStatus,
	selectDrawingToEditStatus
} = drawingSlice.selectors;
