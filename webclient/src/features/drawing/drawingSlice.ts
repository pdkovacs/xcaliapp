import type { PayloadAction } from "@reduxjs/toolkit";
import { createAppSlice } from "../../app/createAppSlice";
import type { Drawing } from "./drawingAPI";
import { fetchDrawing, fetchDrawingList, saveDrawing } from "./drawingAPI";

type AsyncOperationState = "idle" | "loading" | "failed";
interface SaveDrawingActionPayload {
	readonly title: string;
	readonly content: string;
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
}

const initialState: DrawingSliceState = {
	drawingList: {
		getList: {
			status: "idle"
		},
		value: []
	},
	drawingInEdit: {
		open: {
			status: "idle"
		},
		save: {
			status: "idle"
		},
		savedDrawing: {
			title: "",
			content: ""
		},
		currentContent: ""
	}
};

// If you are not using async thunks you can use the standalone `createSlice`.
export const drawingSlice = createAppSlice({
	name: "drawing",
	// `createSlice` will infer the state type from the `initialState` argument
	initialState,
	// The `reducers` field lets us define reducers and generate associated actions
	reducers: create => ({
		getDrawingList: create.asyncThunk(
			async () => {
				const response = await fetchDrawingList();
				// The value we return becomes the `fulfilled` action payload
				return response;
			},
			{
				pending: state => {
					state.drawingList.getList.status = "loading";
				},
				fulfilled: (state, action) => {
					state.drawingList.getList.status = "idle";
					state.drawingList.value = action.payload;
				},
				rejected: state => {
					state.drawingList.getList.status = "failed";
				}
			}
		),
		getDrawingContent: create.asyncThunk(
			async (title: string) => {
				const response = await fetchDrawing(title);
				// The value we return becomes the `fulfilled` action payload
				return {
					title,
					content: response.data
				};
			},
			{
				pending: state => {
					state.drawingInEdit.open.status = "loading";
				},
				fulfilled: (state, action) => {
					state.drawingInEdit.open.status = "idle";
					state.drawingInEdit.savedDrawing = action.payload;
					state.drawingInEdit.currentContent = JSON.parse(action.payload.content);
				},
				rejected: state => {
					state.drawingInEdit.open.status = "failed";
				}
			}
		),
		drawingContentChanged: create.reducer((state, action: PayloadAction<string>) => {
			state.drawingInEdit.currentContent = action.payload;
		}),
		saveDrawingContent: create.asyncThunk(
			async (payload: SaveDrawingActionPayload) => {
				const response = await saveDrawing(payload.title, payload.content);
				// The value we return becomes the `fulfilled` action payload
				return response;
			},
			{
				pending: state => {
					state.drawingInEdit.save.status = "loading";
				},
				fulfilled: (state) => {
					state.drawingInEdit.save.status = "idle";
				},
				rejected: state => {
					state.drawingInEdit.save.status = "failed";
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
