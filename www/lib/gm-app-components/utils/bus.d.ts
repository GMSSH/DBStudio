type EventHandler<T> = (data: T) => void;
export declare class EventBus<T> {
    private eventHandlers;
    on(eventName: string, handler: EventHandler<T>): void;
    off(eventName: string, handler: EventHandler<T>): void;
    emit(eventName: string, data?: T): void;
    clear(eventName?: string): void;
}
export declare const bus: EventBus<any>;
export default bus;
