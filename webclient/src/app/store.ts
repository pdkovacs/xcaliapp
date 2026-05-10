import type { Action, ThunkAction } from "@reduxjs/toolkit";
import { combineSlices, configureStore, createListenerMiddleware } from "@reduxjs/toolkit";
import { setupListeners } from "@reduxjs/toolkit/query";
import { drawingSlice, setSavedDrawing } from "../features/drawing/drawingSlice";
import { drawingApi } from "../features/drawing/drawingApi";
import { drawingErrorLabels } from "../features/drawing/drawingErrors";
import { registerNotificationListeners } from "../features/notifications/notifications";
import { setLocation } from "../utils/set-location";

const rootReducer = combineSlices(drawingSlice, drawingApi);
export type RootState = ReturnType<typeof rootReducer>;

const listenerMiddleware = createListenerMiddleware();

listenerMiddleware.startListening({
	matcher: drawingApi.endpoints.getDrawing.matchFulfilled,
	effect: (action, api) => {
		const arg = action.meta.arg.originalArgs;
		const state = api.getState() as RootState;
		const repos = drawingApi.endpoints.getDrawingRepositories.select()(state).data ?? [];
		const label = repos.find(r => r.name === arg.repoId)?.label ?? "???";
		api.dispatch(setSavedDrawing({
			...action.payload,
			id: arg.drawingId,
			repo: { name: arg.repoId, label }
		}));
		setLocation(arg);
	}
});

listenerMiddleware.startListening({
	matcher: drawingApi.endpoints.createDrawing.matchFulfilled,
	effect: (action, api) => {
		const drawing = action.meta.arg.originalArgs;
		const newId = action.payload;
		api.dispatch(setSavedDrawing({ ...drawing, id: newId }));
		setLocation({ repoId: drawing.repo.name, drawingId: newId });
	}
});

registerNotificationListeners(listenerMiddleware, drawingErrorLabels);

export const makeStore = (preloadedState?: Partial<RootState>) => {
	const store = configureStore({
		reducer: rootReducer,
		middleware: gDM => gDM().prepend(listenerMiddleware.middleware).concat(drawingApi.middleware),
		preloadedState
	});
	setupListeners(store.dispatch);
	return store;
};

export const store = makeStore();

export type AppStore = typeof store;
export type AppDispatch = AppStore["dispatch"];
export type AppThunk<ThunkReturnType = void> = ThunkAction<
	ThunkReturnType,
	RootState,
	unknown,
	Action
>;
