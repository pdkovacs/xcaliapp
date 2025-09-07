import { isEmpty } from "lodash";
import { createAppSlice } from "../../app/createAppSlice";
import { emptyArray } from "../../utils/empty-array";
import type { Drawing, DrawingLists, DrawingRepoRef } from "./drawingAPI";
import { fetchDrawing, fetchDrawingList, createDrawing as createDrawingApi, saveDrawing } from "./drawingAPI";

export enum AsyncOperationState {
	idle = "idle",
	inProgress = "inProgress",
	failed = "failed"
};

const emptyDrawing: Drawing = {
	id: "",
	title: "",
	elements: emptyArray
};

export interface DrawingSliceState {
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

interface LocationParams {
	readonly drawingId?: string;
}

const setLocation = (locationParams: LocationParams) => {
	if (!isEmpty(locationParams.drawingId)) {
		const drawingId = locationParams.drawingId!;
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
			async (id: string) => {
				const response = await fetchDrawing(id);
				setLocation({ drawingId: id });
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
								id: action.meta.arg
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
			setLocation({});
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
				setLocation({ drawingId });
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
		selectDrawingRepos: drawing => {
			const drawingLists = drawing.drawingLists.value;
			const repos: DrawingRepoRef[] = Object.keys(drawingLists).map(repoName => drawingLists[repoName].repoRef);
			return repos;
		},
		selectDrawingLists: drawing => drawing.drawingLists.value,
		selectDrawingListStatus: drawing => drawing.drawingLists.getList.status,
		selectSavedDrawing: drawing => drawing.drawingInEdit.savedDrawing,
		selectDrawingToEditStatus: drawing => drawing.drawingInEdit.open.status,
		selectSaveDrawingStatus: drawing => drawing.drawingInEdit.save.status
	}
});

export const { getDrawingLists, getDrawingContent, clearCanvas, createDrawing, saveDrawingContent } =
	drawingSlice.actions;

export const {
	selectDrawingRepos,
	selectDrawingLists,
	selectDrawingListStatus,
	selectSavedDrawing,
	selectSaveDrawingStatus,
	selectDrawingToEditStatus
} = drawingSlice.selectors;
