import { createAppSlice } from "../../app/createAppSlice";
import type { Drawing, DrawingListItem } from "./drawingAPI";
import { fetchDrawing, fetchDrawingList, createDrawing as createDrawingApi, saveDrawing } from "./drawingAPI";

export enum AsyncOperationState {
	idle = "idle",
	inProgress = "inProgress",
	failed = "failed"
};

export interface DrawingSliceState {
	drawingList: {
		getList: {
			status: AsyncOperationState;
		}
		value: DrawingListItem[];
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
	drawingList: {
		getList: {
			status: AsyncOperationState.idle
		},
		value: []
	},

	drawingInEdit: {
		open: {
			status: AsyncOperationState.idle
		},
		save: {
			status: AsyncOperationState.idle
		},
		savedDrawing: {
			id: "",
			title: "",
			elements: []
		}
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

// If you are not using async thunks you can use the standalone `createSlice`.
export const drawingSlice = createAppSlice({
	name: "drawing",
	initialState,
	reducers: create => ({
		getDrawingList: create.asyncThunk(
			async () => {
				const response = await fetchDrawingList();
				// The value we return becomes the `fulfilled` action payload
				return response;
			},
			{
				pending: state => {
					state.drawingList.getList.status = AsyncOperationState.inProgress;
				},
				fulfilled: (state, action) => {
					state.drawingList.getList.status = AsyncOperationState.idle;
					state.drawingList.value = action.payload;
				},
				rejected: state => {
					state.drawingList.getList.status = AsyncOperationState.failed;
				}
			}
		),
		getDrawingContent: create.asyncThunk(
			async (id: string) => {
				const response = await fetchDrawing(id);
				const desiredPathname = `/drawings/${id}`;
				if (window.location.pathname !== desiredPathname) {
					window.history.pushState({}, id, desiredPathname);
				}
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
							savedDrawing: action.payload,
							currentElements: action.payload.elements
						}
					};
				},
				rejected: state => {
					state.drawingInEdit.open.status = AsyncOperationState.failed;
				}
			}
		),
		createDrawing: create.asyncThunk(
			async (payload: Drawing) => {
				return await createDrawingApi(payload.title, payload.elements);
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
		selectDrawingList: drawing => drawing.drawingList.value,
		selectDrawingListStatus: drawing => drawing.drawingList.getList.status,
		selectSavedDrawing: drawing => drawing.drawingInEdit.savedDrawing,
		selectDrawingToEditStatus: drawing => drawing.drawingInEdit.open.status,
		selectSaveDrawingStatus: drawing => drawing.drawingInEdit.save.status
	}
});

export const { getDrawingList, getDrawingContent, createDrawing, saveDrawingContent } =
	drawingSlice.actions;

export const {
	selectDrawingList,
	selectDrawingListStatus,
	selectSavedDrawing,
	selectSaveDrawingStatus,
	selectDrawingToEditStatus
} = drawingSlice.selectors;
