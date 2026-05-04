import type { Action, ThunkAction } from "@reduxjs/toolkit";
import { combineSlices, configureStore, createListenerMiddleware } from "@reduxjs/toolkit";
import { setupListeners } from "@reduxjs/toolkit/query";
import { drawingSlice, drawingErrorLabels, getDrawingContent } from "../features/drawing/drawingSlice";
import { registerNotificationListeners } from "../features/notifications/notifications";
import { setLocation } from "../utils/set-location";

const rootReducer = combineSlices(drawingSlice);
export type RootState = ReturnType<typeof rootReducer>;

const listenerMiddleware = createListenerMiddleware();

listenerMiddleware.startListening({
	actionCreator: getDrawingContent.fulfilled,
	effect: (action) => setLocation(action.meta.arg)
});

registerNotificationListeners(listenerMiddleware, {
	...drawingErrorLabels
});


export const makeStore = (preloadedState?: Partial<RootState>) => {
	const store = configureStore({
		reducer: rootReducer,
		middleware: gDM => gDM().prepend(listenerMiddleware.middleware),
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
