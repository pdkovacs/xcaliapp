import type { PayloadAction } from "@reduxjs/toolkit";
import { createAppSlice } from "../../app/createAppSlice";
import type { Drawing } from "./drawingAPI";
import { fetchDrawing, fetchDrawingList, plainToBase64, saveDrawing } from "./drawingAPI";

export enum AsyncOperationState {
	idle = "idle",
	inProgress = "inProgress",
	failed = "failed"
};

interface SaveDrawingActionPayload {
	readonly title: string;
	readonly content: string;
}

interface RenameDrawingActionPayload {
	readonly from: string;
	readonly to: string;
}

export interface DrawingSliceState {
	drawingList: {
		getList: {
			status: AsyncOperationState;
		}
		value: string[];
	};
	drawingInEdit: {
		open: {
			status: AsyncOperationState;
		}
		save: {
			status: AsyncOperationState;
		};
		savedDrawing: Drawing;
		currentContent: string;
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
			title: "",
			content: ""
		},
		currentContent: ""
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
			async (title: string) => {
				const response = await fetchDrawing(title);
				const desiredPathname = `/drawings/${plainToBase64(title)}`;
				if (window.location.pathname !== desiredPathname) {
					window.history.pushState({}, title, desiredPathname);
				}
				// The value we return becomes the `fulfilled` action payload
				return {
					title,
					content: response.data
				};
			},
			{
				pending: state => {
					state.drawingInEdit.open.status = AsyncOperationState.inProgress;
				},
				fulfilled: (state, action) => {
					state.drawingInEdit.open.status = AsyncOperationState.idle;
					state.drawingInEdit.savedDrawing = action.payload;
					state.drawingInEdit.currentContent = JSON.parse(action.payload.content);
				},
				rejected: state => {
					state.drawingInEdit.open.status = AsyncOperationState.failed;
				}
			}
		),
		drawingContentChanged: create.reducer((state, action: PayloadAction<string>) => {
			state.drawingInEdit.currentContent = action.payload;
		}),
		saveDrawingContent: create.asyncThunk(
			async (payload: SaveDrawingActionPayload) => {
				await saveDrawing(payload.title, payload.content);
			},
			{
				pending: state => {
					state.drawingInEdit.save.status = AsyncOperationState.inProgress;
				},
				fulfilled: (state, action) => {
					state.drawingInEdit.savedDrawing = action.meta.arg;
					state.drawingInEdit.save.status = AsyncOperationState.idle;
				},
				rejected: state => {
					state.drawingInEdit.save.status = AsyncOperationState.failed;
				}
			}
		),
		renameDrawing: create.reducer((state, action: PayloadAction<RenameDrawingActionPayload>) => {
			if (state.drawingInEdit.savedDrawing.title === action.payload.from) {
				state.drawingInEdit.savedDrawing.title = action.payload.to;
			}
		})
	}),
	// You can define your selectors here. These selectors receive the slice
	// state as their first argument.
	selectors: {
		selectDrawingList: drawing => drawing.drawingList.value,
		selectDrawingListStatus: drawing => drawing.drawingList.getList.status,
		selectSavedDrawing: drawing => drawing.drawingInEdit.savedDrawing,
		selectDrawingToEditStatus: drawing => drawing.drawingInEdit.open.status,
		selectCurrentDrawingContent: drawing => drawing.drawingInEdit.currentContent,
		selectSaveDrawingStatus: drawing => drawing.drawingInEdit.save.status
	}
});

// Action creators are generated for each case reducer function.
export const { getDrawingList, getDrawingContent, drawingContentChanged, saveDrawingContent } =
	drawingSlice.actions;

// Selectors returned by `slice.selectors` take the root state as their first argument.
export const { selectDrawingList,
	selectDrawingListStatus,
	selectSavedDrawing,
	selectSaveDrawingStatus,
	selectCurrentDrawingContent,
	selectDrawingToEditStatus
} = drawingSlice.selectors;
